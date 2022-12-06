const std = @import("std");
const Allocator = std.mem.Allocator;

const ParseCursor = @import("./parse.zig").Cursor;

test "example" {
    const example =
        \\    [D]    
        \\[N] [C]    
        \\[Z] [M] [P]
        \\ 1   2   3 
        \\
        \\move 1 from 2 to 1
        \\move 3 from 1 to 3
        \\move 2 from 2 to 1
        \\move 1 from 1 to 2
        \\
    ;

    const expected =
        \\- NDP
        \\- DCP
        \\-  CD
        \\- C D
        \\> MCD
        \\
    ;

    const allocator = std.testing.allocator;

    var input = std.io.fixedBufferStream(example);
    var output = std.ArrayList(u8).init(allocator);
    defer output.deinit();

    run(allocator, &input, &output) catch |err| {
        std.debug.print("```pre-error output:\n{s}\n```", .{output.items});
        return err;
    };
    try std.testing.expectEqualStrings(expected, output.items);
}

const Stack = std.TailQueue(u8);

const ScenParseError = error{
    UnexpectedLeader,
    UnexpectedEnd,
    UnexpectedBlankLine,
    UnexpectedInputAfterCheck,
    MissingDelim,
    MalformedCrate,
    BadColumnCount,
    BadColumnDigit,
    BadEmptyColumn,
    ColumnCountOutOfOrder,
    ColumnCountMismatch,
};

/// Allocates n-chunks of T elements, which are then kept in an internal singly
/// linked list to be mass destroyed after all such elements are no longer needed.
/// - caller is responsible for any element reuse
/// - only aggregate destruction is supported; no per-element destruction
/// - usage pattern is similar to an arena, but type specific:
///
///     const Nodes = SlabChain(struct {
///         next: ?*@This() = null,
///         // node data fields to taste
///     }, 32).init(yourAllocator);
///     defer Nodes.deinit();
///
///     // build some graph of nodes or whatever
///     var nodeA = Nodes.create();
///     var nodeB = Nodes.create();
///     nodeA.next = nodeB;
fn SlabChain(comptime T: type, comptime n: usize) type {
    const Chunk = struct {
        free: usize = 0,
        chunk: [n]T = undefined,
        prior: ?*@This() = null,

        pub fn create(self: *@This()) ?*T {
            if (self.free >= self.chunk.len) return null;
            const node = &self.chunk[self.free];
            self.free += 1;
            return node;
        }
    };

    return struct {
        allocator: Allocator,
        last: ?*Chunk = null,

        pub fn init(allocator: Allocator) @This() {
            return .{ .allocator = allocator };
        }

        pub fn deinit(self: *@This()) void {
            var last = self.last;
            self.last = null;
            while (last) |slab| {
                last = slab.prior;
                self.allocator.destroy(slab);
            }
        }

        pub fn create(self: *@This()) !*T {
            while (true) {
                if (self.last) |last|
                    if (last.create()) |item|
                        return item;
                var new = try self.allocator.create(Chunk);
                new.* = .{ .prior = self.last };
                self.last = new;
            }
        }
    };
}

const Scene = struct {
    const NodeSlab = SlabChain(Stack.Node, 32);

    nodes: NodeSlab,
    stacks: [9]Stack = [_]Stack{.{}} ** 9,
    used: usize = 0,

    pub fn deinit(self: *@This()) void {
        self.nodes.deinit();
    }

    pub fn Parse(allocator: Allocator, str: []const u8) !@This() {
        var col_count: usize = 0;
        var col_counted: usize = 0;

        var self = @This(){
            .nodes = NodeSlab.init(allocator),
        };

        var lines = std.mem.split(u8, str, "\n");
        while (lines.next()) |line| {
            if (line.len == 0) {
                if (col_counted == 0) return ScenParseError.UnexpectedBlankLine;
                self.used = col_counted;
                return self;
            } else if (col_counted > 0) {
                return ScenParseError.UnexpectedInputAfterCheck;
            }

            var cur = ParseCursor.make(line);

            var col: usize = 0;
            while (true) {
                const leader = cur.consume() orelse break;
                switch (leader) {
                    '[' => {
                        const crate = cur.consume() orelse return ScenParseError.UnexpectedEnd;
                        try self.prepend(col, crate);
                        try cur.expect(']', ScenParseError.MalformedCrate);
                    },

                    ' ' => {
                        const colno = cur.consume() orelse return ScenParseError.UnexpectedEnd;

                        switch (colno) {
                            '1', '2', '3', '4', '5', '6', '7', '8', '9' => {
                                try cur.expect(' ', ScenParseError.BadColumnCount);
                                const check = colno - '1';
                                if (check != col) return ScenParseError.ColumnCountOutOfOrder;
                                col_counted = col + 1;
                            },
                            ' ' => {
                                try cur.expect(' ', ScenParseError.BadEmptyColumn);
                            },
                            else => return ScenParseError.BadColumnDigit,
                        }
                    },

                    else => return ScenParseError.UnexpectedLeader,
                }

                col += 1;

                try cur.expectOrEnd(' ', ScenParseError.MissingDelim);
            }

            if (col_counted > 0) {
                if (col_count != col_counted)
                    return ScenParseError.ColumnCountMismatch;
            } else if (col > col_count) col_count = col;
        }

        return self;
    }

    pub fn top(self: @This(), col: usize) ?u8 {
        if (col < self.used) {
            if (self.stacks[col].last) |node|
                return node.data;
        }
        return null;
    }

    pub fn collectTops(self: @This(), into: []u8) []u8 {
        var col: usize = 0;
        while (col < self.used and col < into.len) : (col += 1) {
            into[col] = if (self.stacks[col].last) |node| node.data else ' ';
        }
        return into[0..col];
    }

    pub fn prepend(self: *@This(), col: usize, crate: u8) !void {
        std.debug.assert(col < 9);
        std.debug.assert(self.used == 0 or col < self.used);
        var node = try self.nodes.create();
        node.data = crate;
        self.stacks[col].prepend(node);
    }

    const MoveError = error{
        InvalidFrom,
        InvalidTo,
        StackEmpty,
    };

    pub fn move(self: *@This(), n: usize, from: usize, to: usize) !void {
        if (from >= 9) return MoveError.InvalidFrom;
        if (to >= 9) return MoveError.InvalidTo;
        const fromStack = &self.stacks[from];
        const toStack = &self.stacks[to];
        if (n == 0) return;

        var subStack = fromStack.last;
        var i: usize = 0;
        while (i < n) {
            if (subStack) |node| {
                i += 1;
                if (i < n) subStack = node.prev;
            } else return MoveError.StackEmpty;
        }

        // TODO: would be nicer to just graft the subStack span directly
        while (subStack) |node| {
            subStack = node.next;
            fromStack.remove(node);
            toStack.append(node);
        }
    }
};

const MoveParseError = error{
    ExpectedMove,
    ExpectedCount,
    ExpectedFrom,
    ExpectedFromID,
    ExpectedTo,
    ExpectedToID,
    UnexpectedTrailer,
};

fn run(
    allocator: Allocator,

    // TODO: better "any .reader()-able / any .writer()-able" interfacing
    input: anytype,
    output: anytype,
) !void {
    var buf = [_]u8{0} ** 4096;
    var in = input.reader();
    var out = output.writer();

    var tmp = std.ArrayList(u8).init(allocator);
    defer tmp.deinit();

    // buffer scene lines, then parse
    while (try in.readUntilDelimiterOrEof(buf[0..], '\n')) |line| {
        if (line.len > 0) {
            try tmp.appendSlice(line);
            try tmp.append('\n');
        } else {
            break;
        }
    }
    var scene = try Scene.Parse(allocator, tmp.items);
    defer scene.deinit();

    var tops: [9]u8 = undefined;

    // interpret moves, change stacks
    while (try in.readUntilDelimiterOrEof(buf[0..], '\n')) |line| {
        try out.print("- {s}\n", .{scene.collectTops(tops[0..])});

        var cur = ParseCursor.make(line);

        try cur.expectStr("move ", MoveParseError.ExpectedMove);
        const n = try cur.expectInt(usize, MoveParseError.ExpectedCount);

        try cur.expectStr(" from ", MoveParseError.ExpectedFrom);
        const from = try cur.expectInt(usize, MoveParseError.ExpectedFromID);

        try cur.expectStr(" to ", MoveParseError.ExpectedTo);
        const to = try cur.expectInt(usize, MoveParseError.ExpectedToID);

        try cur.expectEnd(MoveParseError.UnexpectedTrailer);

        try scene.move(n, from - 1, to - 1);
    }

    try out.print("> {s}\n", .{scene.collectTops(tops[0..])});
}

pub fn main() !void {
    var arena = std.heap.ArenaAllocator.init(std.heap.page_allocator);
    defer arena.deinit();

    const allocator = arena.allocator();

    var input = std.io.getStdIn();
    var output = std.io.getStdOut();

    var bufin = std.io.bufferedReader(input.reader());
    var bufout = std.io.bufferedWriter(output.writer());

    try run(allocator, &bufin, &bufout);
    try bufout.flush();

    // TODO: argument parsing to steer input selection

    // TODO: sentinel-buffered output writer to flush lines progressively

    // TODO: input, output, and run-time metrics
}
