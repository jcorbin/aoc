const std = @import("std");

const mem = std.mem;

const Allocator = mem.Allocator;

test "example" {
    const example_input =
        \\>>><<><>><<<>><>>><<<>>><<<><<<>><>><<>>
        \\
    ;

    const test_cases = [_]struct {
        input: []const u8,
        expected: []const u8,
        config: Config,
        skip: bool = false,
    }{

        // Part 1 example: first 2 rocks trace
        .{
            .config = .{
                .verbose = 2,
                .rocks = 2,
            },
            .input = example_input,
            .expected = 
            \\# T0
            \\    |..@@@@.|
            \\    |.......|
            \\    |.......|
            \\    |.......|
            \\    +-------+
            \\
            \\# T1
            \\    |...@@@@|
            \\    |.......|
            \\    |.......|
            \\    |.......|
            \\    +-------+
            \\
            \\# T2
            \\    |...@@@@|
            \\    |.......|
            \\    |.......|
            \\    +-------+
            \\
            \\# T3
            \\    |...@@@@|
            \\    |.......|
            \\    |.......|
            \\    +-------+
            \\
            \\# T4
            \\    |...@@@@|
            \\    |.......|
            \\    +-------+
            \\
            \\# T5
            \\    |...@@@@|
            \\    |.......|
            \\    +-------+
            \\
            \\# T6
            \\    |...@@@@|
            \\    +-------+
            \\
            \\# T7
            \\    |..@@@@.|
            \\    +-------+
            \\
            \\# T8
            \\    |..####.|
            \\    +-------+
            \\
            \\# T9
            \\    |...@...|
            \\    |..@@@..|
            \\    |...@...|
            \\    |.......|
            \\    |.......|
            \\    |.......|
            \\    |..####.|
            \\    +-------+
            \\
            \\# T10
            \\    |..@....|
            \\    |.@@@...|
            \\    |..@....|
            \\    |.......|
            \\    |.......|
            \\    |.......|
            \\    |..####.|
            \\    +-------+
            \\
            \\# T11
            \\    |..@....|
            \\    |.@@@...|
            \\    |..@....|
            \\    |.......|
            \\    |.......|
            \\    |..####.|
            \\    +-------+
            \\
            \\# T12
            \\    |...@...|
            \\    |..@@@..|
            \\    |...@...|
            \\    |.......|
            \\    |.......|
            \\    |..####.|
            \\    +-------+
            \\
            \\# T13
            \\    |...@...|
            \\    |..@@@..|
            \\    |...@...|
            \\    |.......|
            \\    |..####.|
            \\    +-------+
            \\
            \\# T14
            \\    |..@....|
            \\    |.@@@...|
            \\    |..@....|
            \\    |.......|
            \\    |..####.|
            \\    +-------+
            \\
            \\# T15
            \\    |..@....|
            \\    |.@@@...|
            \\    |..@....|
            \\    |..####.|
            \\    +-------+
            \\
            \\# T16
            \\    |...@...|
            \\    |..@@@..|
            \\    |...@...|
            \\    |..####.|
            \\    +-------+
            \\
            \\# T17
            \\    |...#...|
            \\    |..###..|
            \\    |...#...|
            \\    |..####.|
            \\    +-------+
            \\
            \\# Solution
            \\> 4
            \\
            ,
        },

        // Part 1 example: first few rocks enter
        // .{
        //     .config = .{
        //         .verbose = 1,
        //         .rocks = 11,
        //     },
        //     .input = example_input,
        //     .expected =
        //     \\
        //     ,
        // },
    };

    const allocator = std.testing.allocator;

    for (test_cases) |tc, i| {
        if (tc.skip) continue;
        std.debug.print(
            \\
            \\Test Case {}
            \\===
            \\
        , .{i});
        var input = std.io.fixedBufferStream(tc.input);
        var output = std.ArrayList(u8).init(allocator);
        defer output.deinit();
        run(allocator, &input, &output, tc.config) catch |err| {
            std.debug.print("```pre-error output:\n{s}\n```\n", .{output.items});
            return err;
        };
        try std.testing.expectEqualStrings(tc.expected, output.items);
    }
}

const Parse = @import("parse.zig");
const Timing = @import("perf.zig").Timing(enum {
    parse,
    parseLine,

    solve,

    report,
    overall,
});

const Config = struct {
    verbose: usize = 0,
    rocks: usize = 0,
};

const Move = enum {
    left,
    right,

    const Self = @This();

    pub fn parse(c: u8) !Self {
        return switch (c) {
            '<' => .left,
            '>' => .right,
            else => error.InvalidMove,
        };
    }
};

const Cell = enum {
    null,
    empty,
    rock,
    piece,
    floor,
    wall,
    corner,

    const Self = @This();

    pub fn parse(c: u8) Self {
        return switch (c) {
            '.' => .empty,
            '#' => .rock,
            '@' => .piece,
            '-' => .floor,
            '|' => .wall,
            '+' => .corner,
            ' ' => .null,
            else => .null,
        };
    }

    pub fn parseLine(line: []Cell, s: []const u8) void {
        for (s) |c, i| line[i] = Self.parse(c);
    }

    pub fn glyph(self: Self) u8 {
        return switch (self) {
            .null => ' ',
            .empty => '.',
            .rock => '#',
            .piece => '@',
            .floor => '-',
            .wall => '|',
            .corner => '+',
        };
    }
};

const Patch = struct {
    stride: usize,
    width: usize,
    data: []Cell,

    const Self = @This();

    pub fn clone(self: Self, allocator: Allocator) !Self {
        return Self{
            .stride = self.stride,
            .width = self.width,
            .data = try allocator.dupe(self.data),
        };
    }

    pub fn height(self: Self) usize {
        const n = self.data.len;
        const m = self.stride;
        const size = n / m;
        return if (n % m == 0) size else size + 1;
    }

    pub fn deinit(self: Self, allocator: Allocator) void {
        allocator.free(self.data);
        self.* = undefined;
    }

    pub fn parseMany(allocator: Allocator, ss: []const []const u8) ![]Self {
        var patches = try allocator.alloc(Self, ss.len);
        errdefer allocator.free(patches);
        for (ss) |s, i|
            patches[i] = Self.parse(allocator, s);
        return patches;
    }

    pub fn parse(allocator: Allocator, s: []const u8) !Self {
        var data = try allocator.alloc(Cell, s.len);
        errdefer allocator.free(data);
        return Self.parseInto(data, s);
    }

    pub fn parseInto(data: []Cell, s: []const u8) !Self {
        const stride = if (std.mem.indexOfScalar(u8, s, '\n')) |i| i + 1 else s.len;
        const width = stride - 1;
        {
            var i = width;
            while (i < s.len and i < data.len) : (i += stride)
                if (s[i] != '\n') return error.IrregularPatchString;
        }
        for (data) |*cell, i|
            cell.* = Cell.parse(s[i]);
        return Self{
            .stride = stride,
            .width = width,
            .data = data,
        };
    }
};

const PatchList = struct {
    const Node = struct {
        patch: Patch,
        next: ?*Node,
        prev: ?*Node,

        pub fn initFrom(allocator: Allocator, patch: Patch) !*Node {
            var node = try allocator.create(Node);
            errdefer allocator.free(node);
            node.* = .{
                .patch = try patch.clone(allocator),
                .next = null,
                .prev = null,
            };
            return node;
        }
    };

    const cont_patch = Patch.parse(std.heap.page_allocator,
        \\ |.......|
        \\ |.......|
        \\ |.......|
        \\ |.......|
        \\ |.......|
    ) catch @compileError("must parse cont_patch");

    const init_patch = Patch.parse(std.heap.page_allocator,
        \\ |.......|
        \\ |.......|
        \\ |.......|
        \\ |.......|
        \\ +-------+
    ) catch @compileError("must parse init_patch");

    allocator: Allocator,
    top: *Node,
    bottom: *Node,

    const Self = @This();

    pub fn init(allocator: Allocator) !Self {
        var node = Node.initFrom(allocator, init_patch);
        return Self{
            .allocator = allocator,
            .top = node,
            .bottom = node,
        };
    }

    pub fn deinit(self: *Self) void {
        var it = self.top;
        while (it) |node| : (it = node.next) {
            node.patch.deinit(self.allocator);
            node.* = undefined;
        }
        self.* = undefined;
    }

    pub fn expand(self: *Self) !void {
        var node = try Node.initFrom(self.allocator, cont_patch);
        node.next = self.top;
        self.top.prev = node;
        self.top = node;
    }

    pub fn height(self: Self) usize {
        var size: usize = 0;
        var it = self.top;

        while (it) |node| : (it = node.next) {
            // TODO maybe factor out patch.iterator()
            const h = node.patch.height();
            const w = node.patch.width;
            var y = 0;
            while (y < h) : (y += 1) {
                var x = 0;
                while (x < w) : (x += 1) {
                    const i = y * node.patch.stride + x;
                    switch (node.patch.data[i]) {
                        .null, .corner, .wall, .empty => {},
                        else => {
                            size += h - y;
                            break;
                        },
                    }
                }
            }
        }

        while (it) |node| : (it = node.next)
            size += node.patch.height();

        return size;
    }

    pub fn move(self: *Self, m: Move) void {
        var it = self.top;
        while (it) |node| : (it = node.next) {
            // TODO factor out patch.rowsIterator()
            const h = node.patch.height();
            const w = node.patch.width;

            var had = false;
            var y = 0;
            while (y < h) : (y += 1) {
                const offset = y * node.patch.stride;
                var any = false;
                switch (m) {
                    .left => {
                        var x = 0;
                        var prior = &node.patch.data[offset + x];
                        x += 1;
                        while (x < w) : (x += 1) {
                            var cur = &node.patch.data[offset + x];
                            if (cur.* == .piece) {
                                any = true;
                                if (prior.* != .empty) break;
                                std.mem.swap(Cell, prior, cur);
                            }
                            prior = cur;
                        }
                    },
                    .right => {
                        var x = w - 1;
                        var prior = &node.patch.data[offset + x];
                        while (x > 0) {
                            x -= 1;
                            var cur = &node.patch.data[offset + x];
                            if (cur.* == .piece) {
                                any = true;
                                if (prior.* != .empty) break;
                                std.mem.swap(Cell, prior, cur);
                            }
                            prior = cur;
                        }
                    },
                }

                if (any) {
                    had = true; // found the piece
                } else if (had) {
                    break; // done with the piece
                }
            }
        }
    }

    pub fn drop(self: *Self) void {
        const Loc = struct {
            node: *Node,
            offset: usize,

            fn row(loc: @This()) []Cell {
                return loc.node.patch.data[loc.offset .. loc.offset + loc.node.patch.stride];
            }

            fn prev(loc: @This()) ?@This() {
                return if (loc.offset > 0) .{
                    .node = loc.node,
                    .offset = loc.offset - loc.node.patch.stride,
                } else if (loc.node.prev) |p| .{
                    .node = p,
                    .offset = p.patch.len - p.patch.stride,
                } else null;
            }
        };

        var it = self.top;

        var hadPiece = false;
        var last: ?Loc = null;

        scan_piece: while (it) |node| : (it = node.next) {
            var offset = 0;
            while (offset < node.patch.data.len) : (offset += node.patch.stride) {
                var row = node.patch.data[offset .. offset + node.patch.width];

                // scan to line after last piece line
                const hasPiece = std.mem.indexOfScalar(Cell, row, .piece) != null;
                if (hasPiece) {
                    hadPiece = true;
                } else if (hadPiece) {
                    last = .{ .node = node, .offset = offset };
                    break :scan_piece;
                }
            }
        }

        while (last) |line| {
            var prior = line.prev() orelse break;
            defer last = prior;

            var fromRow = prior.row();
            var toRow = line.row();

            var blocked = false;
            var hasPiece = false;
            for (fromRow) |fromCell, i| {
                if (fromCell == .piece) {
                    hasPiece = true;
                    if (toRow[i] != .empty) {
                        blocked = true;
                        break;
                    }
                }
            }

            if (!hasPiece) break; // done moving piece

            if (blocked) {
                for (fromRow) |*fromCell| {
                    if (fromCell.* == .piece) fromCell.* = .rock;
                }
                // NOTE conveys blockage into prior row (now rock)
            } else {
                for (fromRow) |*fromCell, i| {
                    if (fromCell.* == .piece)
                        std.mem.swap(Cell, fromCell, &toRow[i]);
                }
                // NOTE conveys empty space into prior row
            }
        }
    }
};

const Builder = struct {
    allocator: Allocator,
    arena: std.heap.ArenaAllocator,
    moves: std.ArrayListUnmanaged(Move),

    const Self = @This();

    pub fn parse(allocator: Allocator, reader: anytype) !World {
        var lines = Parse.lineScanner(reader);
        var builder = init: {
            var cur = try lines.next() orelse return error.NoInput;
            break :init Builder.initLine(allocator, cur) catch |err| return cur.carp(err);
        };
        // var lineTime = try timing.timer(.parseLine);
        while (try lines.next()) |cur| {
            builder.parseLine(cur) catch |err| return cur.carp(err);
            // TODO bring back: try lineTime.lap();
        }
        return builder.finish();
    }

    pub fn initLine(allocator: Allocator, cur: *Parse.Cursor) !Self {
        var self = Self{
            .allocator = allocator,
            .arena = std.heap.ArenaAllocator.init(allocator),
        };
        try self.parseLine(cur);
        return self;
    }

    pub fn parseLine(self: *Self, cur: *Parse.Cursor) !void {
        try self.moves.ensureUnusedCapacity(self.arena.allocator(), cur.buf.len);
        while (cur.i < cur.buf.len) : (cur.i += 1) {
            switch (cur.buf[cur.i]) {
                '\n' => {},
                else => |c| self.moves.appendAssumeCapacity(try Move.parse(c)),
            }
        }
        if (cur.live()) return error.ExtraINput;
    }

    pub fn finish(self: *Self) World {
        return World{
            .allocator = self.allocator,
            .arena = self.arena,
            .moves = self.moves.items,
        };
    }
};

const pieces = Patch.parseMany(std.heap.page_allocator,
    \\ ####
,
    \\  # 
    \\ ###
    \\  # 
,
    \\   #
    \\   #
    \\ ###
,
    \\ #
    \\ #
    \\ #
    \\ #
,
    \\ ##
    \\ ##
) catch @compileError("must parse pieces");

const World = struct {
    allocator: Allocator,
    arena: std.heap.ArenaAllocator,
    moves: []Move,

    const Self = @This();

    pub fn deinit(self: *Self) void {
        self.arena.deinit();
    }
};

fn run(
    allocator: Allocator,

    // TODO: better "any .reader()-able / any .writer()-able" interfacing
    input: anytype,
    output: anytype,
    config: Config,
) !void {
    var timing = try Timing.start(allocator);
    defer timing.deinit();
    defer timing.printDebugReport();

    var out = output.writer();

    // FIXME: hookup your config
    _ = config;

    var arena = std.heap.ArenaAllocator.init(allocator);
    defer arena.deinit();

    var world = try Builder.parse(allocator, input.reader());
    defer world.deinit();
    try timing.markPhase(.parse);

    // FIXME: solve...
    try timing.markPhase(.solve);

    try out.print(
        \\# Solution
        \\> {}
        \\
    , .{
        42,
    });
    try timing.markPhase(.report);

    try timing.finish(.overall);
}

const ArgParser = @import("args.zig").Parser;

const MainAllocator = std.heap.GeneralPurposeAllocator(.{
    .enable_memory_limit = true,
    .verbose_log = false,
});

var gpa = MainAllocator{};

var log_level: std.log.Level = .warn;

pub fn log(comptime level: std.log.Level, comptime scope: @TypeOf(.EnumLiteral), comptime format: []const u8, args: anytype) void {
    if (@enumToInt(level) > @enumToInt(log_level)) return;
    const prefix = "[" ++ comptime level.asText() ++ "] (" ++ @tagName(scope) ++ "): ";
    std.debug.getStderrMutex().lock();
    defer std.debug.getStderrMutex().unlock();
    const stderr = std.io.getStdErr().writer();
    nosuspend stderr.print(prefix ++ format ++ "\n", args) catch return;
}

pub fn main() !void {
    gpa.setRequestedMemoryLimit(4 * 1024 * 1024 * 1024);

    var allocator = gpa.allocator();

    var input = std.io.getStdIn();
    var output = std.io.getStdOut();
    var config = Config{};
    var bufferOutput = true;

    {
        var argsArena = std.heap.ArenaAllocator.init(allocator);
        defer argsArena.deinit();

        var args = try ArgParser.init(argsArena.allocator());
        defer args.deinit();

        while (args.next()) |arg| {
            if (!arg.isOption()) {
                var prior = input;
                input = try std.fs.cwd().openFile(arg.have, .{});
                prior.close();
                std.log.info("reading input from {s}", .{arg.have});
            } else if (arg.is(.{ "-h", "--help" })) {
                std.debug.print(
                    \\Usage: {s} [options] [INPUT_FILE]
                    \\
                    \\Options:
                    \\
                    \\  -v or
                    \\  --verbose
                    \\    print world state after evaluating each input line
                    \\
                    \\  --raw-output
                    \\    don't buffer stdout writes
                    \\
                , .{args.progName()});
                std.process.exit(0);
            } else if (arg.is(.{ "-v", "--verbose" })) {
                if (@enumToInt(log_level) < @enumToInt(std.log.Level.debug))
                    log_level = @intToEnum(std.log.Level, @enumToInt(log_level) + 1);
                config.verbose += 1;
            } else if (arg.is(.{"--raw-output"})) {
                bufferOutput = false;
            } else return error.InvalidArgument;
        }
    }

    (do_run: {
        var bufin = std.io.bufferedReader(input.reader());

        if (!bufferOutput)
            break :do_run run(allocator, &bufin, output, config);

        var bufout = std.io.bufferedWriter(output.writer());
        defer bufout.flush() catch {};

        break :do_run run(allocator, &bufin, &bufout, config);
        // TODO: sentinel-buffered output writer to flush lines progressively
        // ... may obviate the desire for raw / non-buffered output else
    }) catch |err| {
        if (err == mem.Allocator.Error.OutOfMemory) {
            const have_leaks = gpa.detectLeaks();
            std.log.err("{} detectLeaks() -> {}", .{ err, have_leaks });
        }
        return err;
    };
    if (gpa.detectLeaks()) return error.LeakedMemory;
}
