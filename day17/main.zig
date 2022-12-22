const std = @import("std");

const assert = std.debug.assert;
const mem = std.mem;

// TODO is there a std.meta.* thing for this?
pub fn Optional(comptime T: type) type {
    return @Type(.{ .Optional = .{ .child = T } });
}

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
            \\Final rock pile is 4 high:
            \\|...#...|
            \\|..###..|
            \\|...#...|
            \\|..####.|
            \\+-------+
            \\
            ,
        },

        // Part 1 example: first few rocks enter
        .{
            .config = .{
                .verbose = 1,
                .rocks = 11,
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
            \\A new rock begins falling:
            \\|..@....|
            \\|..@....|
            \\|..@....|
            \\|..@....|
            \\|.......|
            \\|.......|
            \\|.......|
            \\|..#....|
            \\|..#....|
            \\|####...|
            \\|..###..|
            \\|...#...|
            \\|..####.|
            \\+-------+
            \\
            \\A new rock begins falling:
            \\|..@@...|
            \\|..@@...|
            \\|.......|
            \\|.......|
            \\|.......|
            \\|....#..|
            \\|..#.#..|
            \\|..#.#..|
            \\|#####..|
            \\|..###..|
            \\|...#...|
            \\|..####.|
            \\+-------+
            \\
            \\A new rock begins falling:
            \\|..@@@@.|
            \\|.......|
            \\|.......|
            \\|.......|
            \\|....##.|
            \\|....##.|
            \\|....#..|
            \\|..#.#..|
            \\|..#.#..|
            \\|#####..|
            \\|..###..|
            \\|...#...|
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
            \\|.####..|
            \\|....##.|
            \\|....##.|
            \\|....#..|
            \\|..#.#..|
            \\|..#.#..|
            \\|#####..|
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
            \\|..#....|
            \\|.###...|
            \\|..#....|
            \\|.####..|
            \\|....##.|
            \\|....##.|
            \\|....#..|
            \\|..#.#..|
            \\|..#.#..|
            \\|#####..|
            \\|..###..|
            \\|...#...|
            \\|..####.|
            \\+-------+
            \\
            \\A new rock begins falling:
            \\|..@....|
            \\|..@....|
            \\|..@....|
            \\|..@....|
            \\|.......|
            \\|.......|
            \\|.......|
            \\|.....#.|
            \\|.....#.|
            \\|..####.|
            \\|.###...|
            \\|..#....|
            \\|.####..|
            \\|....##.|
            \\|....##.|
            \\|....#..|
            \\|..#.#..|
            \\|..#.#..|
            \\|#####..|
            \\|..###..|
            \\|...#...|
            \\|..####.|
            \\+-------+
            \\
            \\A new rock begins falling:
            \\|..@@...|
            \\|..@@...|
            \\|.......|
            \\|.......|
            \\|.......|
            \\|....#..|
            \\|....#..|
            \\|....##.|
            \\|....##.|
            \\|..####.|
            \\|.###...|
            \\|..#....|
            \\|.####..|
            \\|....##.|
            \\|....##.|
            \\|....#..|
            \\|..#.#..|
            \\|..#.#..|
            \\|#####..|
            \\|..###..|
            \\|...#...|
            \\|..####.|
            \\+-------+
            \\
            \\A new rock begins falling:
            \\|..@@@@.|
            \\|.......|
            \\|.......|
            \\|.......|
            \\|....#..|
            \\|....#..|
            \\|....##.|
            \\|##..##.|
            \\|######.|
            \\|.###...|
            \\|..#....|
            \\|.####..|
            \\|....##.|
            \\|....##.|
            \\|....#..|
            \\|..#.#..|
            \\|..#.#..|
            \\|#####..|
            \\|..###..|
            \\|...#...|
            \\|..####.|
            \\+-------+
            \\
            \\Final rock pile is 18 high:
            \\|...####|
            \\|....#..|
            \\|....#..|
            \\|....##.|
            \\|##..##.|
            \\|######.|
            \\|.###...|
            \\|..#....|
            \\|.####..|
            \\|....##.|
            \\|....##.|
            \\|....#..|
            \\|..#.#..|
            \\|..#.#..|
            \\|#####..|
            \\|..###..|
            \\|...#...|
            \\|..####.|
            \\+-------+
            \\
            ,
        },

        // Part 1 example: final answer
        .{
            .config = .{
                .rocks = 2022,
            },
            .input = example_input,
            .expected = 
            \\Final rock pile will be 3068 high.
            \\
            ,
        },

        // Part 2 example: final answer
        .{
            .config = .{
                .rocks = 1_000_000_000_000,
            },
            .input = example_input,
            .expected = 
            \\Final rock pile will be 1514285714288 high.
            \\
            ,
        },
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

const Timing = @import("perf.zig").Timing(enum {
    parse,
    parseLine,

    simulateRockBatch,
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

    pub fn isAllEmpty(row: []const Cell) bool {
        switch (row[0]) {
            .wall, .empty => {},
            else => return false,
        }
        switch (row[row.len - 1]) {
            .wall, .empty => {},
            else => return false,
        }
        return std.mem.allEqual(Cell, row[1 .. row.len - 1], .empty);
    }

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

    pub fn format(self: Self, comptime _: []const u8, _: std.fmt.FormatOptions, writer: anytype) !void {
        try writer.writeByte(self.glyph());
    }
};

const Patch = struct {
    stride: usize,
    width: usize,
    data: []Cell,

    const Self = @This();

    pub fn format(self: Self, comptime how: []const u8, _: std.fmt.FormatOptions, writer: anytype) !void {
        if (std.mem.eql(u8, how, "d")) {
            return writer.print("Patch{{ .width = {}, .height() = {} }}", .{ self.width, self.height() });
        }
        try writer.print("Patch{{ .stride = {}, .width = {}, .data = {any} }}", .{
            self.stride,
            self.width,
            self.data,
        });
    }

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
            patches[i] = Self.parse(allocator, s) catch |err| {
                std.log.warn("unable to parse `{s}`", .{
                    std.fmt.fmtSliceEscapeLower(s),
                });
                return err;
            };
        return patches;
    }

    pub fn parse(allocator: Allocator, s: []const u8) !Self {
        var data = try allocator.alloc(Cell, s.len);
        errdefer allocator.free(data);
        return Self.parseInto(data, s);
    }

    pub fn parseInto(data: []Cell, s: []const u8) !Self {
        const stride = validate: {
            if (std.mem.indexOfScalar(u8, s, '\n')) |i| {
                const stride = i + 1;
                var j = i;
                while (j < s.len and j < data.len) : (j += stride)
                    if (s[j] != '\n') return error.IrregularPatchString;
                break :validate stride;
            }
            break :validate s.len + 1;
        };

        const width = stride - 1;
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
        next: ?*Node = null,
        prev: ?*Node = null,

        pub fn format(self: *const Node, comptime _: []const u8, _: std.fmt.FormatOptions, writer: anytype) !void {
            try writer.print("{*}{{ .prev = {*}, .next = {*}, .patch = {d} }}", .{
                self,
                self.prev,
                self.next,
                self.patch,
            });
        }

        pub fn initFrom(allocator: Allocator, patch: Patch) !*Node {
            var node = try allocator.create(Node);
            errdefer allocator.destroy(node);
            node.* = .{ .patch = try patch.clone(allocator) };
            return node;
        }
    };

    pub fn Loc(comptime is_const: bool) type {
        return struct {
            node: if (is_const) *const Node else *Node,
            offset: usize = 0,

            fn row(loc: @This()) if (is_const) []const Cell else []Cell {
                return loc.node.patch.data[loc.offset .. loc.offset + loc.node.patch.width];
            }

            fn height(loc: @This()) usize {
                var size =
                    loc.node.patch.height() -
                    loc.offset / loc.node.patch.stride;
                var it = loc.node.next;
                while (it) |node| : (it = node.next)
                    size += node.patch.height();
                return size;
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
                    .offset = p.patch.data.len - p.patch.stride + 1,
                } else null;
            }
        };
    }

    top: ?*Node = null,
    len: usize = 0,
    height: usize = 0,

    const Self = @This();

    pub fn deinit(self: *Self, allocator: Allocator) void {
        var it = self.top;
        while (it) |node| : (it = node.next) {
            assert(node.next != node);
            node.patch.deinit(allocator);
            node.* = undefined;
        }
        self.* = undefined;
    }

    pub fn format(self: Self, comptime _: []const u8, _: std.fmt.FormatOptions, writer: anytype) !void {
        var at = self.topLoc();

        // skip all-empty header lines
        while (at) |loc| : (at = loc.next()) {
            if (!Cell.isAllEmpty(loc.row())) break;
        }

        var first = true;
        while (at) |loc| : (at = loc.next()) {
            if (first) first = false else try writer.writeByte('\n');
            const row = loc.row();
            for (row) |c|
                try writer.writeByte(c.glyph());
        }
    }

    pub fn expand(self: *Self, allocator: Allocator, patch: Patch) !void {
        if (self.top) |prior| {
            if (patch.width != prior.patch.width) return error.InvalidPatchWidth;
        }
        var node = try Node.initFrom(allocator, patch);
        if (self.top) |top| {
            node.next = top;
            top.prev = node;
        }
        self.top = node;
        self.len += 1;
        self.height += patch.height();
        // TODO wen compact
    }

    pub fn availHeight(self: Self) usize {
        assert(self.height >= self.usedHeight());
        return self.height - self.usedHeight();
    }

    pub fn usedHeight(self: Self) usize {
        var empty = @as(usize, 0);
        var at = self.topLoc();
        row: while (at) |loc| : (at = loc.next()) {
            if (Cell.isAllEmpty(loc.row())) {
                empty += 1;
            } else break :row;
        }
        return self.height - empty;
    }

    pub fn topLoc(self: Self) ?Loc(true) {
        return .{ .node = self.top orelse return null };
    }

    pub fn topLocMut(self: *Self) ?Loc(false) {
        return .{ .node = self.top orelse return null };
    }

    pub fn move(self: *Self, m: Move) bool {
        var at = self.topLocMut();

        var start: Optional(@TypeOf(at)) = null;
        while (at) |loc| : (at = loc.next()) {
            const row = loc.row();

            const any = scan: {
                switch (m) {
                    .left => {
                        var x = @as(usize, 0);
                        var prior = row[x];
                        x += 1;
                        while (x < row.len) : (x += 1) {
                            const cur = row[x];
                            if (cur == .piece) {
                                if (prior != .empty) return false; // blocked
                                break :scan true;
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
                                break :scan true;
                            }
                            prior = cur;
                        }
                    },
                }
                break :scan false;
            };

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
            var at = self.topLocMut();
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
        if (last == null) return false;

        // check if any piece line is blocked
        const blocked = check_piece: {
            var check = last;
            while (check) |line| {
                var prior = line.prev() orelse break;
                defer check = prior;

                const fromRow = prior.row();
                const toRow = line.row();

                var hasPiece = false;
                for (fromRow) |fromCell, i| {
                    if (fromCell == .piece) switch (toRow[i]) {
                        .empty, .piece => hasPiece = true,
                        else => break :check_piece true,
                    };
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

    pub fn initParse(allocator: Allocator, reader: anytype) !World {
        var builder = Builder.init(allocator);
        try builder.parse(reader);

        return try builder.finish();
    }

    pub fn parse(self: *Self, reader: anytype) !void {
        while (true)
            switch (reader.readByte() catch |err| switch (err) {
                error.EndOfStream => break,
                else => return err,
            }) {
                '\n' => {},
                else => |c| try self.moves.append(self.arena.allocator(), try Move.parse(c)),
            };
    }

    pub fn init(allocator: Allocator) Self {
        return .{
            .allocator = allocator,
            .arena = std.heap.ArenaAllocator.init(allocator),
        };
    }

    pub fn finish(self: *Self) !World {
        const cont_patch = try Patch.parse(self.arena.allocator(),
            \\|.......|
            \\|.......|
            \\|.......|
            \\|.......|
            \\|.......|
        );

        const init_patch = try Patch.parse(self.arena.allocator(),
            \\|.......|
            \\|.......|
            \\|.......|
            \\|.......|
            \\+-------+
        );

        const pieces = try Patch.parseMany(self.arena.allocator(), split: {
            var chunks = std.mem.split(u8,
                \\@@@@
                \\
                \\ @ 
                \\@@@
                \\ @ 
                \\
                \\  @
                \\  @
                \\@@@
                \\
                \\@
                \\@
                \\@
                \\@
                \\
                \\@@
                \\@@
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
            .init_patch = init_patch,
            .cont_patch = cont_patch,
            .moves = self.moves.items,
        };
    }
};

const World = struct {
    allocator: Allocator,
    arena: std.heap.ArenaAllocator,

    pieces: []Patch,
    init_patch: Patch,
    cont_patch: Patch,

    moves: []Move,
    room: PatchList = .{},

    const Self = @This();

    pub fn deinit(self: *Self) void {
        self.arena.deinit();
    }

    pub fn movesIterator(self: Self) struct {
        moves: []Move,
        i: usize = 0,

        pub fn next(it: *@This()) Move {
            defer it.i += 1;
            return it.moves[it.i % it.moves.len];
        }
    } {
        return .{ .moves = self.moves };
    }

    pub fn ensureRoomHeight(self: *Self, needed: usize) !void {
        while (true) {
            const prior = self.room.availHeight();
            if (prior >= needed) return;

            const add_patch = if (prior == 0) self.init_patch else self.cont_patch;
            try self.room.expand(self.arena.allocator(), add_patch);
            assert(self.room.availHeight() > prior);
        }
    }

    pub fn addPiece(
        self: *Self,
        piece: Patch,
        opts: struct {
            offset: usize = 1,
            space: usize = 1,
        },
    ) !void {
        const space = piece.height() + opts.space;
        try self.ensureRoomHeight(space);
        assert(piece.width < self.room.top.?.patch.width - opts.offset);

        // defer {
        //     std.debug.print("ADDED PIECE {}x{} {}\n", .{ piece.width, piece.height(), opts });
        //     var y = @as(usize, 0);
        //     var at = self.room.topLoc();
        //     while (at) |loc| : (at = loc.next()) {
        //         std.debug.print("  y={} : `{any}`\n", .{ y, loc.row() });
        //         y += 1;
        //     }
        // }

        var into = find: {
            var at = self.room.topLocMut();
            while (at) |loc| : (at = loc.next()) {
                if (!Cell.isAllEmpty(loc.row())) {
                    var i = @as(usize, 0);
                    while (i < space) : (i += 1) at = at.?.prev();
                    break :find at;
                }
            }
            break :find null;
        } orelse return error.NoPlaceToAdd;

        var off = @as(usize, 0);
        while (off < piece.data.len) : (off += piece.stride) {
            var row = into.row();

            var x = @as(usize, 0);
            while (x < piece.width) : (x += 1) {
                switch (piece.data[off + x]) {
                    .null => {},
                    else => |c| row[opts.offset + x] = c,
                }
            }

            into = into.next() orelse return error.NoPlaceToAdd;
        }
    }
};

fn gcd(a: u64, b: u64) u64 {
    return if (b == 0) a else gcd(b, a % b);
}

fn lcm(a: u64, b: u64) u64 {
    return if (a > b)
        (a / gcd(a, b)) * b
    else
        (b / gcd(a, b)) * a;
}

const status_log = std.log.scoped(.status);

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

    var world = try Builder.initParse(allocator, input.reader());
    defer world.deinit();
    try timing.markPhase(.parse);

    const final_height = simulate: {
        defer timing.markPhase(.simulate) catch {};

        const rep_rocks = lcm(world.moves.len, world.pieces.len);

        var extra: std.MultiArrayList(struct {
            height: usize,
            delta: usize,
        }) = .{};
        const extra_space = 100 * rep_rocks;

        if (config.rocks > rep_rocks) {
            try extra.ensureUnusedCapacity(arena.allocator(), extra_space);
        }

        const sim_rocks = config.rocks;

        var moves = world.movesIterator();
        var rock_i = @as(usize, 0);
        var rockTime = try timing.timer(.simulateRockBatch);
        var elapsed = @as(u64, 0);
        const frocks = @intToFloat(f64, extra_space);
        while (rock_i < sim_rocks) : (rock_i += 1) {
            if (extra.capacity > 0 and rock_i > 0) {
                if (extra.len >= extra.capacity) {
                    std.log.warn("haven't found a cycle in {} rocks ( {} rounds )", .{
                        extra.len,
                        extra.len / rep_rocks,
                    });
                    try extra.ensureUnusedCapacity(arena.allocator(), extra_space);
                }

                const last_round = if (extra.len >= rep_rocks) extra.len / rep_rocks else null;

                const last_i = if (last_round) |round| round * rep_rocks - 1 else null;
                // const last_i = if (extra.len > 0) extra.len - 1 else null;

                const last = if (last_i) |i| extra.get(i).height else 0;

                const height = world.room.usedHeight() - 1;
                const delta = height - last;

                extra.appendAssumeCapacity(.{
                    .height = height,
                    .delta = delta,
                });

                if (rock_i % rep_rocks == 0) {
                    if (cycle: {
                        const deltas = extra.items(.delta);
                        if (deltas.len < 2 * rep_rocks) break :cycle null;

                        // main phase: search successive powers of two
                        const len = find: {
                            var power = @as(usize, 1);
                            var len = @as(usize, 1);
                            var tortoise_i = @as(usize, 0);
                            var hare_i = @as(usize, 1);
                            while (deltas[(1 + tortoise_i) * rep_rocks - 1] != deltas[(1 + hare_i) * rep_rocks - 1] or len <= 2) {
                                if (power == len) { // time to start a new power of two?
                                    tortoise_i = hare_i;
                                    power *= 2;
                                    len = 0;
                                }
                                hare_i += 1;
                                len += 1;
                                if (hare_i * rep_rocks >= deltas.len) break :cycle null;
                            }
                            break :find len;
                        };

                        // Find the position of the first repetition of such length
                        const at = measure: {
                            var at = @as(usize, 0);
                            var tortoise_i = @as(usize, 0);
                            var hare_i = len;
                            // hare and tortoise move at same speed until they agree
                            while (deltas[(1 + tortoise_i) * rep_rocks - 1] != deltas[(1 + hare_i) * rep_rocks - 1]) {
                                tortoise_i += 1;
                                hare_i += 1;
                                at += 1;
                                if (hare_i * rep_rocks >= deltas.len) break :cycle null;
                            }
                            break :measure at;
                        };

                        break :cycle struct { at: usize, len: usize }{ .at = at, .len = len };
                    }) |cycle| {
                        const heights = extra.items(.height);

                        const cycle_height = accum: {
                            const deltas = extra.items(.delta);

                            // compute
                            var cycle_height = @as(usize, 0);
                            var i = @as(usize, 0);
                            while (i < cycle.len) : (i += 1) {
                                const j = cycle.at + i;
                                const k = (1 + j) * rep_rocks - 1;
                                cycle_height += deltas[k];
                            }

                            // validate
                            const first_i = cycle.at + cycle.len - 1;
                            const second_i = cycle.at + 2 * cycle.len - 1;

                            const first_k = (1 + first_i) * rep_rocks - 1;
                            const second_k = (1 + second_i) * rep_rocks - 1;

                            const first = heights[first_k];
                            const second = heights[second_k];
                            if (second - first != cycle_height) {
                                std.log.err("cycle at {} len {} height {} != {}@<{},{}> - {}@<{},{}>", .{
                                    cycle.at,
                                    cycle.len,
                                    cycle_height,

                                    second,
                                    second_i,
                                    second_k,

                                    first,
                                    first_i,
                                    first_k,
                                });

                                // XXX [error] (default): cycle at 66 len 32 height 2473759 != 10049758@<129,6559149> - 7575962@<97,4944589>
                                //
                                // XXX cycle at 66 len 32
                                // XXX height 2473759
                                // XXX != 10049758@<129,6559149> - 7575962@<97,4944589>
                                // XXX 10049758 - 7575962
                                // XXX 2473796
                                //
                                // XXX a 2473759
                                // XXX b 2473796
                                //
                                // XXX 2473759 - 2473796
                                // XXX -37

                                i = 0;
                                for ([_]usize{ 1, 2 }) |n| {
                                    while (i < n * cycle.len) : (i += 1) {
                                        const j = cycle.at + i;
                                        const k = (1 + j) * rep_rocks - 1;
                                        std.log.err("cycle {} @<{},{}> {}", .{ n, j, k, extra.get(k) });
                                    }
                                }

                                return error.CycleHeightInvalid;
                            }

                            break :accum cycle_height;
                        };

                        const predicted_height = fast_forward: {
                            const cycle_rocks = cycle.len * rep_rocks;
                            const end_i = cycle.at + 2 * cycle.len - 1;
                            const end_k = (1 + end_i) * rep_rocks - 1;
                            const end_height = heights[end_k];

                            var k = end_k;
                            var h = end_height;
                            while (k < sim_rocks) {
                                k += cycle_rocks;
                                h += cycle_height;
                            }

                            const back_k = k + 1 - sim_rocks;
                            var i = @as(usize, 0);
                            while (i < back_k) : (i += 1) {
                                const back_i = back_k - i;
                                const back_delta =
                                    heights[heights.len - back_i] -
                                    heights[heights.len - back_i - 1];
                                h -= back_delta;
                            }
                            break :fast_forward h;
                        };

                        break :simulate predicted_height + 1; // plus floor
                    }
                }
            }

            if (rock_i > 0 and rock_i % 1_000 == 0) {
                const t = rockTime.lap();
                const f = @intToFloat(f64, rock_i);
                const p = f / frocks;
                elapsed += t;
                const per = @intToFloat(f64, elapsed) / f;
                const rem = (frocks - f) * per;
                status_log.info("{d:.2}% dropped 1000 rocks in {}, eta: {}", .{
                    p * 100.0,
                    std.fmt.fmtDuration(t),
                    std.fmt.fmtDuration(@floatToInt(u64, @round(rem))),
                });
            }

            try world.addPiece(
                world.pieces[rock_i % world.pieces.len],
                // Each rock appears so that its left edge is two units away from
                // the left wall and its bottom edge is three units above the
                // highest rock in the room (or the floor, if there isn't one).
                .{
                    .offset = 3, // wall + 2 empties
                    .space = 3,
                },
            );
            if (config.verbose > 0) {
                if (rock_i == 0)
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

            var step = @as(usize, 0);
            const stepLimit = world.room.height;

            while (true) {
                step += 1;
                if (step > stepLimit) return error.TickLimitExceeded;

                const move = moves.next();

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

                if (!didDrop) break;
            }
        }
        break :simulate world.room.usedHeight();
    };

    switch (final_height) {
        0 => try out.print(
            \\No rocks, empty world.
            \\
        , .{}),
        else => |h| {
            const height = h - 1; // floor discount
            if (h > world.room.usedHeight())
                try out.print(
                    \\Final rock pile will be {} high.
                    \\
                , .{height})
            else if (config.verbose > 0)
                try out.print(
                    \\Final rock pile is {} high:
                    \\{}
                    \\
                , .{ height, world.room })
            else
                try out.print(
                    \\Final rock pile is {} high.
                    \\
                , .{height});
        },
    }
    try timing.markPhase(.report);

    try timing.finish(.overall);
}

const ArgParser = @import("args.zig").Parser;

const MainAllocator = std.heap.GeneralPurposeAllocator(.{
    .enable_memory_limit = true,
    .verbose_log = false,
});

var gpa = MainAllocator{};

pub const log_level = std.log.Level.debug; // NOTE this just causes all log sites to compile in
var actual_log_level = std.log.Level.warn; // NOTE this is the one that matters, may be increased at runtime

pub fn log(comptime level: std.log.Level, comptime scope: @TypeOf(.EnumLiteral), comptime format: []const u8, args: anytype) void {
    switch (scope) {
        .perf, .status => {},
        else => if (@enumToInt(level) > @enumToInt(actual_log_level)) return,
    }

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
                    \\  -r COUNT or
                    \\  --rocks COUNT
                    \\    how many rocks to drop
                    \\
                    \\  -v or
                    \\  --verbose
                    \\    * increases amount of stdout world state and
                    \\    * increases logging level each time given
                    \\
                    \\  --raw-output
                    \\    don't buffer stdout writes
                    \\
                , .{args.progName()});
                std.process.exit(0);
            } else if (arg.is(.{ "-r", "--rocks" })) {
                var count_arg = args.next() orelse return error.MissingQueryLine;
                config.rocks = try count_arg.parseInt(usize, 10);
            } else if (arg.is(.{ "-v", "--verbose" })) {
                if (@enumToInt(actual_log_level) < @enumToInt(std.log.Level.debug))
                    actual_log_level = @intToEnum(std.log.Level, @enumToInt(actual_log_level) + 1);
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
