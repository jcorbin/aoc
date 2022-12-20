const std = @import("std");

const assert = std.debug.assert;
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
            \\The first rock begins falling:
            \\|..@@@@.|
            \\|.......|
            \\|.......|
            \\|.......|
            \\+-------+
            \\
            \\Jet of gas pushes rock right:
            \\|...@@@@|
            \\|.......|
            \\|.......|
            \\|.......|
            \\+-------+
            \\
            \\Rock falls 1 unit:
            \\|...@@@@|
            \\|.......|
            \\|.......|
            \\+-------+
            \\
            \\Jet of gas pushes rock right, but nothing happens:
            \\|...@@@@|
            \\|.......|
            \\|.......|
            \\+-------+
            \\
            \\Rock falls 1 unit:
            \\|...@@@@|
            \\|.......|
            \\+-------+
            \\
            \\Jet of gas pushes rock right, but nothing happens:
            \\|...@@@@|
            \\|.......|
            \\+-------+
            \\
            \\Rock falls 1 unit:
            \\|...@@@@|
            \\+-------+
            \\
            \\Jet of gas pushes rock left:
            \\|..@@@@.|
            \\+-------+
            \\
            \\Rock falls 1 unit, causing it to come to rest:
            \\|..####.|
            \\+-------+
            \\
            \\A new rock begins falling:
            \\|...@...|
            \\|..@@@..|
            \\|...@...|
            \\|.......|
            \\|.......|
            \\|.......|
            \\|..####.|
            \\+-------+
            \\
            \\Jet of gas pushes rock left:
            \\|..@....|
            \\|.@@@...|
            \\|..@....|
            \\|.......|
            \\|.......|
            \\|.......|
            \\|..####.|
            \\+-------+
            \\
            \\Rock falls 1 unit:
            \\|..@....|
            \\|.@@@...|
            \\|..@....|
            \\|.......|
            \\|.......|
            \\|..####.|
            \\+-------+
            \\
            \\Jet of gas pushes rock right:
            \\|...@...|
            \\|..@@@..|
            \\|...@...|
            \\|.......|
            \\|.......|
            \\|..####.|
            \\+-------+
            \\
            \\Rock falls 1 unit:
            \\|...@...|
            \\|..@@@..|
            \\|...@...|
            \\|.......|
            \\|..####.|
            \\+-------+
            \\
            \\Jet of gas pushes rock left:
            \\|..@....|
            \\|.@@@...|
            \\|..@....|
            \\|.......|
            \\|..####.|
            \\+-------+
            \\
            \\Rock falls 1 unit:
            \\|..@....|
            \\|.@@@...|
            \\|..@....|
            \\|..####.|
            \\+-------+
            \\
            \\Jet of gas pushes rock right:
            \\|...@...|
            \\|..@@@..|
            \\|...@...|
            \\|..####.|
            \\+-------+
            \\
            \\Rock falls 1 unit, causing it to come to rest:
            \\|...#...|
            \\|..###..|
            \\|...#...|
            \\|..####.|
            \\+-------+
            \\
            \\A new rock begins falling:
            \\|....@..|
            \\|....@..|
            \\|..@@@..|
            \\|.......|
            \\|.......|
            \\|.......|
            \\|...#...|
            \\|..###..|
            \\|...#...|
            \\|..####.|
            \\+-------+
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

    simulateRockStep,
    simulateRock,
    simulate,

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

    pub fn name(self: Self) []const u8 {
        return switch (self) {
            .left => "left",
            .right => "right",
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
            .data = try allocator.dupe(Cell, self.data),
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
            patches[i] = try Self.parse(allocator, s);
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
            errdefer allocator.destroy(node);
            node.* = .{
                .patch = try patch.clone(allocator),
                .next = null,
                .prev = null,
            };
            return node;
        }
    };

    const Loc = struct {
        node: *Node,
        offset: usize = 0,

        fn row(loc: @This()) []Cell {
            return loc.node.patch.data[loc.offset .. loc.offset + loc.node.patch.width];
        }

        fn next(loc: @This()) ?@This() {
            const offset = loc.offset + loc.node.patch.stride;
            return if (offset < loc.node.patch.data.len) .{
                .node = loc.node,
                .offset = offset,
            } else if (loc.node.next) |n| .{
                .node = n,
                .offset = 0,
            } else null;
        }

        fn prev(loc: @This()) ?@This() {
            return if (loc.offset > 0) .{
                .node = loc.node,
                .offset = loc.offset - loc.node.patch.stride,
            } else if (loc.node.prev) |p| .{
                .node = p,
                .offset = p.patch.data.len - p.patch.stride,
            } else null;
        }
    };

    allocator: Allocator,
    top: *Node,
    bottom: *Node,

    const Self = @This();

    pub fn init(allocator: Allocator, init_patch: Patch) !Self {
        var node = try Node.initFrom(allocator, init_patch);
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

    pub fn expand(self: *Self, patch: Patch) !void {
        if (patch.width != self.top.patch.width) return error.InvalidPatchWidth;
        var node = try Node.initFrom(self.allocator, patch);
        node.next = self.top;
        self.top.prev = node;
        self.top = node;
    }

    pub fn availHeight(self: Self) usize {
        return self.height() - self.usedHeight();
    }

    pub fn height(self: Self) usize {
        var size = @as(usize, 0);
        var it: ?*Node = self.top;
        while (it) |node| : (it = node.next)
            size += node.patch.height();
        return size;
    }

    pub fn usedHeight(self: Self) usize {
        var size = @as(usize, 0);
        var it: ?*Node = self.top;

        while (it) |node| : (it = node.next) {
            // TODO maybe factor out patch.iterator()
            const h = node.patch.height();
            const w = node.patch.width;
            var y = @as(usize, 0);
            while (y < h) : (y += 1) {
                var x = @as(usize, 0);
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

    pub fn move(self: *Self, m: Move) bool {
        var at: ?Loc = Loc{ .node = self.top };
        var start: ?Loc = null;
        while (at) |loc| : (at = loc.next()) {
            const row = loc.row();

            var any = false;
            switch (m) {
                .left => {
                    var x = @as(usize, 0);
                    var prior = row[x];
                    x += 1;
                    while (x < row.len) : (x += 1) {
                        const cur = row[x];
                        if (cur == .piece) {
                            if (prior != .empty) return false; // blocked
                            any = true;
                        }
                        prior = cur;
                    }
                },
                .right => {
                    var x = row.len - 1;
                    var prior = row[x];
                    while (x > 0) {
                        x -= 1;
                        const cur = row[x];
                        if (cur == .piece) {
                            if (prior != .empty) return false; // blocked
                            any = true;
                        }
                        prior = cur;
                    }
                },
            }

            if (any) {
                if (start == null) start = loc; // found the piece
            } else if (start != null) break; // done with the piece
        }

        at = start orelse return false; // no piece to move

        // now we can just move the piece, swapping its cells for empty space left/right
        while (at) |loc| : (at = loc.next()) {
            var row = loc.row();
            if (std.mem.indexOfScalar(Cell, row, .piece) == null) break; // done with the piece

            switch (m) {
                .left => {
                    var x = @as(usize, 0);
                    var prior = &row[x];
                    x += 1;
                    while (x < row.len) : (x += 1) {
                        var cur = &row[x];
                        if (cur.* == .piece) {
                            assert(prior.* == .empty);
                            std.mem.swap(Cell, prior, cur);
                        }
                        prior = cur;
                    }
                },
                .right => {
                    var x = row.len - 1;
                    var prior = &row[x];
                    while (x > 0) {
                        x -= 1;
                        var cur = &row[x];
                        if (cur.* == .piece) {
                            assert(prior.* == .empty);
                            std.mem.swap(Cell, prior, cur);
                        }
                        prior = cur;
                    }
                },
            }
        }

        return true;
    }

    pub fn drop(self: *Self) bool {
        // scan to line after last piece line
        var last = scan_piece: {
            var hadPiece = false;
            var at: ?Loc = Loc{ .node = self.top };
            while (at) |loc| : (at = loc.next()) {
                const offset = loc.offset;
                const patch = loc.node.patch;
                const row = patch.data[offset .. offset + patch.width];
                const hasPiece = std.mem.indexOfScalar(Cell, row, .piece) != null;
                if (hasPiece) {
                    hadPiece = true;
                } else if (hadPiece) {
                    break :scan_piece at;
                }
            }
            break :scan_piece null;
        };

        // check if any piece line is blocked
        const blocked = check_piece: {
            var check = last;
            while (check) |line| {
                var prior = line.prev() orelse break;
                defer last = prior;
                const fromRow = prior.row();
                const toRow = line.row();
                var hasPiece = false;
                for (fromRow) |fromCell, i| {
                    if (fromCell == .piece) {
                        if (toRow[i] != .empty) break :check_piece true;
                        hasPiece = true;
                    }
                }
                if (!hasPiece) break;
            }
            break :check_piece false;
        };

        if (blocked) {
            // convert to rock
            while (last) |line| {
                var prior = line.prev() orelse break;
                defer last = prior;

                var fromRow = prior.row();
                if (std.mem.indexOfScalar(Cell, fromRow, .piece) == null) break;

                for (fromRow) |*fromCell| {
                    if (fromCell.* == .piece) fromCell.* = .rock;
                }
            }
            return false;
        } else {
            // move down (swap with empty space checked above);
            while (last) |line| {
                var prior = line.prev() orelse break;
                defer last = prior;

                var fromRow = prior.row();
                if (std.mem.indexOfScalar(Cell, fromRow, .piece) == null) break;
                var toRow = line.row();

                for (fromRow) |*fromCell, i| {
                    if (fromCell.* == .piece) {
                        assert(toRow[i] == .empty);
                        std.mem.swap(Cell, fromCell, &toRow[i]);
                    }
                }
            }
            return true;
        }
    }
};

const Builder = struct {
    allocator: Allocator,
    arena: std.heap.ArenaAllocator,
    moves: std.ArrayListUnmanaged(Move) = .{},

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
        return try builder.finish();
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

    pub fn finish(self: *Self) !World {
        const cont_patch = try Patch.parse(self.arena.allocator(),
            \\ |.......|
            \\ |.......|
            \\ |.......|
            \\ |.......|
            \\ |.......|
        );

        const init_patch = try Patch.parse(self.arena.allocator(),
            \\ |.......|
            \\ |.......|
            \\ |.......|
            \\ |.......|
            \\ +-------+
        );

        const pieces = try Patch.parseMany(self.arena.allocator(), split: {
            var chunks = std.mem.split(u8,
                \\ ####
                \\
                \\  # 
                \\ ###
                \\  # 
                \\
                \\   #
                \\   #
                \\ ###
                \\
                \\ #
                \\ #
                \\ #
                \\ #
                \\
                \\ ##
                \\ ##
            , "\n\n");
            var parts = try std.ArrayList([]const u8).initCapacity(self.arena.allocator(), count: {
                var n = @as(usize, 0);
                while (chunks.next()) |_| n += 1;
                break :count n;
            });
            chunks.reset();
            while (chunks.next()) |part| parts.appendAssumeCapacity(part);
            break :split parts.items;
        });

        return World{
            .allocator = self.allocator,
            .arena = self.arena,
            .pieces = pieces,
            .cont_patch = cont_patch,
            .moves = self.moves.items,
            .room = try PatchList.init(self.arena.allocator(), init_patch),
        };
    }
};

const World = struct {
    allocator: Allocator,
    arena: std.heap.ArenaAllocator,

    pieces: []Patch,
    cont_patch: Patch,

    moves: []Move,
    room: PatchList,

    const Self = @This();

    pub fn deinit(self: *Self) void {
        self.arena.deinit();
    }

    pub fn movesIterator(self: Self) struct {
        moves: []Move,
        i: usize = 0,

        pub fn next(it: *@This()) ?Move {
            if (it.i < it.moves.len) {
                defer it.i += 1;
                return it.moves[it.i];
            }
            return null;
        }
    } {
        return .{ .moves = self.moves };
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

    var arena = std.heap.ArenaAllocator.init(allocator);
    defer arena.deinit();

    var world = try Builder.parse(allocator, input.reader());
    defer world.deinit();
    try timing.markPhase(.parse);

    var moves = world.movesIterator();
    var rock_i = @as(usize, 0);
    var rockTime = try timing.timer(.simulateRock);
    while (rock_i < config.rocks) : (rock_i += 1) {
        defer rockTime.lap() catch {};
        const piece = world.pieces[rock_i % world.pieces.len];

        // Each rock appears so that its left edge is two units away from the
        // left wall and its bottom edge is three units above the highest rock
        // in the room (or the floor, if there isn't one).

        if (world.room.availHeight() < 3)
            try world.room.expand(world.cont_patch);
        assert(world.room.availHeight() >= 3);

        {
            var into = world.room.top.patch;
            assert(piece.width < into.width - 3);
            assert(piece.height() < into.height());

            var off = @as(usize, 0);
            while (off < piece.data.len) : (off += piece.stride) {
                var at = off + 3; // wall + 2 empties
                var x = @as(usize, 0);
                while (x < piece.width) : (x += 1)
                    into.data[at + x] = piece.data[off + x];
            }
        }

        if (config.verbose > 0) {
            if (config.verbose < 2)
                try out.print(
                    \\{}
                    \\
                    \\
                , .{world.room})
            else if (rock_i == 0)
                try out.print(
                    \\The first rock begins falling:
                    \\{}
                    \\
                    \\
                , .{world.room})
            else
                try out.print(
                    \\A new rock begins falling:
                    \\{}
                    \\
                    \\
                , .{world.room});
        }

        var stepTime = try timing.timer(.simulateRock);
        while (true) {
            defer stepTime.lap() catch {};

            const move = moves.next() orelse return error.OutOfMoves;

            const didMove = world.room.move(move);
            if (config.verbose > 1) {
                if (didMove)
                    try out.print(
                        \\Jet of gas pushes rock {s}:
                        \\{}
                        \\
                        \\
                    , .{ move.name(), world.room })
                else
                    try out.print(
                        \\Jet of gas pushes rock {s}, but nothing happens:
                        \\{}
                        \\
                        \\
                    , .{ move.name(), world.room });
            }

            const didDrop = world.room.drop();
            if (config.verbose > 1) {
                if (didDrop)
                    try out.print(
                        \\Rock falls 1 unit:
                        \\{}
                        \\
                        \\
                    , .{world.room})
                else
                    try out.print(
                        \\Rock falls 1 unit, causing it to come to rest:
                        \\{}
                        \\
                        \\
                    , .{world.room});
            }

            if (didDrop) break;
        }
    }

    try timing.markPhase(.simulate);

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
