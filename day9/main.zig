const std = @import("std");
const Allocator = std.mem.Allocator;

test "example" {
    const example =
        \\R 4
        \\U 4
        \\L 3
        \\D 1
        \\R 4
        \\D 1
        \\L 5
        \\R 2
        \\
    ;

    const expected =
        \\# Eval 1. R 4
        \\    ......
        \\    ......
        \\    ......
        \\    ......
        \\    s..TH.
        \\
        \\# Eval 2. U 4
        \\    ....H.
        \\    ....T.
        \\    ......
        \\    ......
        \\    s.....
        \\
        \\# Eval 3. L 3
        \\    .HT...
        \\    ......
        \\    ......
        \\    ......
        \\    s.....
        \\
        \\# Eval 4. D 1
        \\    ..T...
        \\    .H....
        \\    ......
        \\    ......
        \\    s.....
        \\
        \\# Eval 5. R 4
        \\    ......
        \\    ....TH
        \\    ......
        \\    ......
        \\    s.....
        \\
        \\# Eval 6. D 1
        \\    ......
        \\    ....T.
        \\    .....H
        \\    ......
        \\    s.....
        \\
        \\# Eval 7. L 5
        \\    ......
        \\    ......
        \\    HT....
        \\    ......
        \\    s.....
        \\
        \\# Eval 8. R 2
        \\    ......
        \\    ......
        \\    .TH...
        \\    ......
        \\    s.....
        \\
        \\# Part 1
        \\    ..##..
        \\    ...##.
        \\    .####.
        \\    ....#.
        \\    s###..
        \\
        \\> 13
        \\
    ;

    const allocator = std.testing.allocator;

    var input = std.io.fixedBufferStream(example);
    var output = std.ArrayList(u8).init(allocator);
    defer output.deinit();

    run(allocator, &input, &output, .{
        .showEvals = true,
        .bounds = .{
            .from = .{ .x = 0, .y = -4 },
            .to = .{ .x = 6, .y = 1 },
        },
    }) catch |err| {
        std.debug.print("```pre-error output:\n{s}\n```\n", .{output.items});
        return err;
    };
    try std.testing.expectEqualStrings(expected, output.items);
}

const Parse = @import("./parse.zig");
const Timing = @import("./perf.zig").Timing;

// TODO really should sort out how to Vector
const Point = struct {
    x: i32,
    y: i32,

    const Self = @This();

    pub fn of(x: i32, y: i32) Point {
        return .{ .x = x, .y = y };
    }

    pub fn add(a: Self, b: Self) Self {
        return .{ .x = a.x + b.x, .y = a.y + b.y };
    }

    pub fn sub(a: Self, b: Self) Self {
        return .{ .x = a.x - b.x, .y = a.y - b.y };
    }

    pub fn absDiff(a: Self, b: Self) struct { x: u32, y: u32 } {
        return .{
            .x = @intCast(u32, if (a.x > b.x) a.x - b.x else b.x - a.x),
            .y = @intCast(u32, if (a.y > b.y) a.y - b.y else b.y - a.y),
        };
    }

    pub fn indexInto(self: Self, offset: usize, stride: usize) usize {
        return offset + stride * @intCast(usize, self.y) + @intCast(usize, self.x);
    }
};

const Rect = struct {
    from: Point,
    to: Point,

    const Self = @This();

    pub fn width(self: Self) u32 {
        const x2 = self.to.x;
        const x1 = self.from.x;
        return if (x2 > x1) @intCast(u32, x2 - x1) else 0;
    }

    pub fn height(self: Self) u32 {
        const y2 = self.to.y;
        const y1 = self.from.y;
        return if (y2 > y1) @intCast(u32, y2 - y1) else 0;
    }

    pub fn contains(self: Self, p: Point) bool {
        const x2 = self.to.x;
        const x1 = self.from.x;
        const y2 = self.to.y;
        const y1 = self.from.y;
        return x1 <= p.x and p.x < x2 and y1 <= p.y and p.y < y2;
    }

    pub fn expandTo(self: *Self, p: Point) void {
        if (self.to.x > self.from.x and self.to.y > self.from.y) {
            if (p.x < self.from.x) self.from.x = p.x;
            if (p.y < self.from.y) self.from.y = p.y;
            if (p.x >= self.to.x) self.to.x = p.x + 1;
            if (p.y >= self.to.y) self.to.y = p.y + 1;
        } else {
            self.from.x = p.x;
            self.from.y = p.y;
            self.to.x = p.x + 1;
            self.to.y = p.y + 1;
        }
    }
};

fn Layer(comptime Datum: type) type {
    return struct {
        const Index = std.AutoHashMap(Point, usize);

        // TODO: evaluate switching to MultiArrayList
        const Data = std.ArrayList(struct {
            loc: Point,
            tag: Datum,
        });

        bounds: Rect = .{
            .from = .{ .x = 0, .y = 0 },
            .to = .{ .x = 0, .y = 0 },
        },
        index: Index,
        data: Data,

        const RegionIterator = struct {
            const It = @This();

            items: Self.Data.Slice,
            bounds: Rect,
            at: Point = .{ .x = 0, .y = 0 },
            i: usize = 0,

            pub fn next(it: *It) ?usize {
                var i = it.i;
                while (i < it.items.len) : (i += 1) {
                    const loc = it.items[i].loc;
                    if (it.bounds.contains(loc)) {
                        it.at = loc;
                        it.i = i + 1;
                        return i;
                    }
                }
                return null;
            }
        };

        const Self = @This();

        pub fn init(allocator: Allocator) Self {
            return .{
                .index = Index.init(allocator),
                .data = Data.init(allocator),
            };
        }

        pub fn deinit(self: *Self) void {
            self.index.deinit();
            self.data.deinit();
        }

        pub fn move(self: *Self, from: Point, to: Point) !void {
            const kv = self.index.fetchRemove(from) orelse return;
            const i = kv.value;
            try self.index.put(to, i);
            self.data.items[i].loc = to;
            self.bounds.expandTo(to); // NOTE: bounds doesn't shrink ex from
        }

        pub fn set(self: *Self, loc: Point, tag: Datum) !void {
            if (self.index.get(loc)) |i| {
                self.data.items[i].tag = tag;
            } else {
                const i = self.data.items.len;
                try self.data.append(.{
                    .loc = loc,
                    .tag = tag,
                });
                try self.index.put(loc, i);

                self.bounds.expandTo(loc);
            }
        }

        pub fn get(self: *Self, loc: Point) ?Datum {
            return if (self.index.get(loc)) |i| self.data.items[i].tag else null;
        }

        pub fn withinRegion(self: *Self, bounds: Rect) RegionIterator {
            return .{
                .items = self.data.items,
                .bounds = bounds,
            };
        }
    };
}

const World = struct {
    const Tag = enum {
        empty,
        start,
        visited,
        head,
        tail,

        pub fn glyph(self: @This()) u8 {
            return switch (self) {
                .empty => '.',
                .start => 's',
                .visited => '#',
                .head => 'H',
                .tail => 'T',
            };
        }
    };

    arena: std.heap.ArenaAllocator,
    tmpString: std.ArrayList(u8),

    background: Layer(Tag),
    headAt: Point = .{ .x = 0, .y = 0 },
    tailAt: Point = .{ .x = 0, .y = 0 },

    const Self = @This();

    pub fn init(allocator: Allocator) Self {
        var self = Self{
            .arena = std.heap.ArenaAllocator.init(allocator),
            .background = Layer(Tag).init(allocator),
            .tmpString = std.ArrayList(u8).init(allocator),
        };
        return self;
    }

    pub fn deinit(self: *Self) void {
        self.tmpString.deinit();
        self.background.deinit();
        self.arena.deinit();
    }

    pub fn startAt(self: *Self, at: Point) !void {
        self.headAt = at;
        self.tailAt = at;
        try self.background.set(self.tailAt, .start);
    }

    pub fn evalMove(self: *Self, cur: *Parse.Cursor) !void {
        const moveCode = cur.consume() orelse return error.MissingMoveCode;
        const moveDir = switch (moveCode) {
            'R' => Point.of(1, 0),
            'L' => Point.of(-1, 0),
            'U' => Point.of(0, -1),
            'D' => Point.of(0, 1),
            else => return error.InvalidMoveCode,
        };

        try cur.expect(' ', error.MissingMoveSpace);

        const moveCount = try cur.expectInt(u8, error.MissingMoveCount);
        if (moveCount < 1) return error.InvalidMoveCount;

        var i: u8 = 0;
        while (i < moveCount) : (i += 1)
            try self.moveHeadBy(moveDir);
    }

    pub fn moveHeadBy(self: *Self, by: Point) !void {
        try self.moveHeadTo(self.headAt.add(by));
    }

    pub fn moveHeadTo(self: *Self, to: Point) !void {
        self.headAt = to;

        // move tail if needed
        //
        // T . H   . T H
        // . . . > . . .
        // . . .   . . .
        //
        // . . H   . . H
        // . . . > . . T
        // . . T   . . .
        //
        // . . H   . T H
        // T . . > . . .
        // . . .   . . .
        //
        // . . H   . . H
        // . . . > . . T
        // . T .   . . .
        const d = self.headAt.sub(self.tailAt);
        const distX = std.math.absInt(d.x) catch 9; // 9 chosen to ensure `else` below
        const distY = std.math.absInt(d.y) catch 9; // ... and because dilbert of course
        if (distX > 1 or distY > 1) {
            if (d.x > 0) {
                self.tailAt.x += 1;
            } else if (d.x < 0) {
                self.tailAt.x -= 1;
            }
            if (d.y > 0) {
                self.tailAt.y += 1;
            } else if (d.y < 0) {
                self.tailAt.y -= 1;
            }
        }

        // mark background
        switch (self.background.get(self.tailAt) orelse Tag.empty) {
            .empty => try self.background.set(self.tailAt, .visited),
            else => {},
        }
    }

    pub fn toString(self: *Self, config: struct {
        prefix: []const u8 = "",
        showStart: bool = false,
        showVisited: bool = false,
        showHeadTail: bool = false,
    }) ![]u8 {
        const bounds = self.background.bounds;
        const offset = config.prefix.len; // no extra sapce at start of line
        const stride = offset + bounds.width() + 1; // space for newline terminator

        const memSize = bounds.height() * stride - 1; // less final newline
        try self.tmpString.resize(memSize);
        var buf = self.tmpString.items;

        // fill empty
        std.mem.set(u8, buf, Tag.empty.glyph());

        { // prefix lines
            var i: usize = 0;
            while (i < buf.len) : (i += stride)
                std.mem.copy(u8, buf[i..], config.prefix);
        }

        { // terminate lines
            var i = stride - 1;
            while (i < buf.len) : (i += stride)
                buf[i] = '\n';
        }

        if (config.showStart or config.showVisited) {
            var it = self.background.withinRegion(bounds);
            while (it.next()) |i| {
                const j = it.at.sub(bounds.from).indexInto(offset, stride);
                const tag = self.background.data.items[i].tag;
                switch (tag) {
                    .start => {
                        if (config.showStart) buf[j] = tag.glyph();
                    },
                    .visited => {
                        if (config.showVisited) buf[j] = tag.glyph();
                    },
                    else => {},
                }
            }
        }

        if (config.showHeadTail) {
            if (bounds.contains(self.tailAt)) buf[
                self.tailAt.sub(bounds.from).indexInto(offset, stride)
            ] = Tag.tail.glyph();
            // head over tail (if same spot)
            if (bounds.contains(self.headAt)) buf[
                self.headAt.sub(bounds.from).indexInto(offset, stride)
            ] = Tag.head.glyph();
        }

        return self.tmpString.items;
    }
};

const Config = struct {
    bounds: Rect = .{
        .from = .{ .x = 0, .y = 0 },
        .to = .{ .x = 0, .y = 0 },
    },
    showEvals: bool = false,
};

fn run(
    allocator: Allocator,

    // TODO: better "any .reader()-able / any .writer()-able" interfacing
    input: anytype,
    output: anytype,
    config: Config,
) !void {
    var timing = try Timing(enum {
        eval,
        evalLine,
        evalShow,
        part1,
        part2,
        overall,
    }).start(allocator);
    defer timing.deinit();
    defer timing.printDebugReport();

    var world = World.init(allocator);
    defer world.deinit();
    world.background.bounds = config.bounds;
    try world.startAt(.{ .x = 0, .y = 0 });

    var lines = Parse.lineScanner(input.reader());
    var out = output.writer();

    // evaluate moves
    while (try lines.next()) |*cur| {
        var lineTime = try std.time.Timer.start();
        world.evalMove(cur) catch |err| {
            std.debug.print("failed to eval line #{}: `{s}`\n", .{ cur.count, cur.buf });
            return err;
        };
        try timing.collect(.evalLine, lineTime.lap());

        if (config.showEvals) {
            try out.print(
                \\# Eval {}. {s}
                \\{s}
                \\
                \\
            , .{
                cur.count,
                cur.buf,
                try world.toString(.{
                    .prefix = "    ",
                    .showHeadTail = true,
                    .showStart = true,
                }),
            });
            try timing.collect(.evalShow, lineTime.lap());
        }
    }
    try timing.markPhase(.eval);

    // Part 1: How many positions does the tail of the rope visit at least once?
    {
        var count: usize = 0;
        for (world.background.data.items) |item| {
            switch (item.tag) {
                .start, .visited => count += 1,
                else => {},
            }
        }

        try out.print(
            \\# Part 1
            \\{s}
            \\
            \\> {}
            \\
        , .{
            try world.toString(.{
                .prefix = "    ",
                .showStart = true,
                .showVisited = true,
            }),
            count,
        });
    }
    try timing.markPhase(.part1);

    // // FIXME: solve part 2
    // {
    //     try out.print("\n# Part 2\n", .{});
    //     try out.print("> {}\n", .{42});
    // }
    // try timing.markPhase(.part2);

    try timing.finish(.overall);
}

pub fn main() !void {
    const allocator = std.heap.page_allocator;

    var input = std.io.getStdIn();
    var output = std.io.getStdOut();

    var bufin = std.io.bufferedReader(input.reader());
    var bufout = std.io.bufferedWriter(output.writer());

    try run(allocator, &bufin, &bufout, .{
        .showEvals = false, // TODO argv flag/option
    });
    try bufout.flush();

    // TODO: argument parsing to steer input selection

    // TODO: sentinel-buffered output writer to flush lines progressively

    // TODO: input, output, and run-time metrics
}
