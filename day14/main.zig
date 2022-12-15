const std = @import("std");
const math = std.math;
const assert = std.debug.assert;

const mem = std.mem;
const Allocator = mem.Allocator;
const Vector = std.meta.Vector;

test "example" {
    const example_input =
        \\498,4 -> 498,6 -> 496,6
        \\503,4 -> 502,4 -> 502,9 -> 494,9
        \\
    ;

    const test_cases = [_]struct {
        input: []const u8,
        expected: []const u8,
        config: Config,
    }{
        // Part 1 example
        .{
            .config = .{
                .verbose = 1,
            },
            .input = example_input,
            .expected = 
            \\# Solution
            // \\@<494,0>
            // \\    ......+...
            // \\    ..........
            // \\    ..........
            // \\    ..........
            // \\    ....#...##
            // \\    ....#...#.
            // \\    ..###...#.
            // \\    ........#.
            // \\    ........#.
            // \\    #########.
            // \\
            // \\@<494,0>
            // \\    ......+...
            // \\    ..........
            // \\    ......o...
            // \\    .....ooo..
            // \\    ....#ooo##
            // \\    ...o#ooo#.
            // \\    ..###ooo#.
            // \\    ....oooo#.
            // \\    .o.ooooo#.
            // \\    #########.
            // \\
            \\@<493,0>
            \\    .......+....
            \\    .......~....
            \\    ......~o....
            \\    .....~ooo...
            \\    ....~#ooo##.
            \\    ...~o#ooo#..
            \\    ..~###ooo#..
            \\    ..~..oooo#..
            \\    .~o.ooooo#..
            \\    ~#########..
            \\    ~...........
            // \\
            \\> 24
            \\
            ,
        },
    };

    const allocator = std.testing.allocator;

    for (test_cases) |tc, i| {
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
};

const Point = Vector(2, i16);
const Size = Vector(2, u16);
const Rect = Vector(4, i16);

const LineIterator = struct {
    from: Point,
    to: Point,
    d: Point,
    i: usize = 0,
    len: usize,

    const Self = @This();

    pub fn next(self: *Self) ?Point {
        if (self.i >= self.len) return null;
        const r = self.from;
        self.from = r + self.d;
        self.i += 1;
        return r;
    }
};

fn pointRange(from: Point, to: Point) LineIterator {
    const delta = to - from;
    return .{
        .from = from,
        .to = to,
        .d = math.sign(delta),
        .len = @intCast(usize, @maximum(
            math.absInt(delta[0]) catch unreachable,
            math.absInt(delta[1]) catch unreachable,
        )),
    };
}

fn rectSize(r: Rect) Size {
    return .{
        if (r[2] > r[0]) @intCast(u16, r[2] - r[0]) else 0,
        if (r[3] > r[1]) @intCast(u16, r[3] - r[1]) else 0,
    };
}

fn rectContains(r: Rect, to: Point) bool {
    return to[0] >= r[0] and
        to[1] >= r[1] and
        to[0] < r[2] and
        to[1] < r[3];
}

fn rectExpand(r: Rect, to: Point) Rect {
    const size = rectSize(r);
    const top = Point{ to[0] + 1, to[1] + 1 };
    if (size[0] * size[1] == 0)
        return .{
            to[0],  to[1],
            top[0], top[1],
        };
    return .{
        @minimum(r[0], to[0]),  @minimum(r[1], to[1]),
        @maximum(r[2], top[0]), @maximum(r[3], top[1]),
    };
}

const Builder = struct {
    const Step = struct {
        loc: Point,
        end: bool = false,
    };
    const Data = std.ArrayListUnmanaged(Step);

    allocator: Allocator,
    data: Data = .{},

    const Self = @This();

    pub fn initLine(allocator: Allocator, cur: *Parse.Cursor) !Self {
        var self = Self{
            .allocator = allocator,
            .data = try Data.initCapacity(allocator, 1024),
        };
        try self.parseLine(cur);
        return self;
    }

    pub fn parseLine(self: *Self, cur: *Parse.Cursor) !void {
        while (true) {
            const x = cur.consumeInt(i16, 10) orelse return error.ExpectedXCoordinate;
            if (!cur.have(',')) return error.ExpectedCoordinateComma;
            const y = cur.consumeInt(i16, 10) orelse return error.ExpectedYCoordinate;
            const loc = Point{ x, y };
            cur.star(' ');

            const i = self.data.items.len;
            try self.data.append(self.allocator, .{ .loc = loc });
            var last = &self.data.items[i];

            if (cur.haveLiteral("->")) {
                cur.star(' ');
                continue;
            }

            if (!cur.live()) {
                last.end = true;
                break;
            }

            return error.ExpectedArrow;
        }
    }

    pub fn finish(self: *Self) !World {
        defer self.data.deinit(self.allocator);

        const source = Point{ 500, 0 };

        var bounds = Rect{ 0, 0, 0, 0 };
        bounds = rectExpand(bounds, source);
        for (self.data.items) |item|
            bounds = rectExpand(bounds, item.loc);

        if (bounds[0] > 0) bounds[0] -= 1;
        if (bounds[1] > 0) bounds[1] -= 1;
        if (bounds[2] < math.maxInt(i16)) bounds[2] += 1;
        if (bounds[3] < math.maxInt(i16)) bounds[3] += 1;

        const size = rectSize(bounds);
        const width = size[0];
        const height = size[1];
        var world = World{
            .allocator = self.allocator,
            .source = source,
            .bounds = bounds,
            .width = width,
            .height = height,
            .data = try self.allocator.alloc(Cell, width * height),
        };
        mem.set(Cell, world.data, .air);
        world.set(source, .source);

        var prior: ?Point = null;
        for (self.data.items) |*step| {
            const to = step.loc;
            if (prior) |from| {
                var line = pointRange(from, to);
                while (line.next()) |loc|
                    world.set(loc, .rock);
            }
            if (step.end) {
                world.set(to, .rock);
                prior = null;
            } else {
                prior = to;
            }
        }

        return world;
    }
};

const Cell = enum {
    air,
    source,
    rock,
    sand,
    mark,

    const Self = @This();

    pub fn glyph(self: Self) u8 {
        return switch (self) {
            .air => '.',
            .source => '+',
            .rock => '#',
            .sand => 'o',
            .mark => '~',
        };
    }
};

const World = struct {
    allocator: Allocator,

    source: Point,
    bounds: Rect,
    width: usize,
    height: usize,
    data: []Cell,

    marked: std.ArrayListUnmanaged(Point) = .{},

    const Self = @This();

    pub fn deinit(self: *Self) void {
        self.data.deinit(self.allocator);
        self.marked.deinit(self.allocator);
    }

    pub fn pour(self: *Self) !bool {
        // unmark any prior
        for (self.marked.items) |at|
            self.set(at, .air);
        self.marked.clearRetainingCapacity();

        var at = self.source + Point{ 0, 1 };
        if (self.get(at) != .air) return error.SourceBlocked;

        try self.marked.ensureTotalCapacity(self.allocator, 4 * (self.width + self.height));

        fall: while (true) {
            for ([_]Point{
                .{ 0, 1 }, // down
                .{ -1, 1 }, // down-left
                .{ 1, 1 }, // down-right
            }) |move| {
                const to = at + move;

                if (!self.isInside(to)) {
                    try self.mark(at);
                    return false;
                }

                switch (self.get(to)) {
                    .rock, .sand, .source => continue,
                    .air, .mark => {
                        try self.mark(at);
                        at = to;
                        continue :fall;
                    },
                }
            }
            self.set(at, .sand);
            return true;
        }
    }

    pub fn mark(self: *Self, at: Point) !void {
        if (self.get(at) != .air) return error.MayOnlyMarkAir;
        try self.marked.append(self.allocator, at);
        self.set(at, .mark);
    }

    pub fn isInside(self: *Self, loc: Point) bool {
        return rectContains(self.bounds, loc);
    }

    pub fn ref(self: *Self, loc: Point) *Cell {
        assert(self.isInside(loc));
        const x = @intCast(usize, loc[0] - self.bounds[0]);
        const y = @intCast(usize, loc[1] - self.bounds[1]);
        const i = self.width * y + x;
        return &self.data[i];
    }

    pub fn set(self: *Self, loc: Point, cell: Cell) void {
        self.ref(loc).* = cell;
    }

    pub fn get(self: *Self, loc: Point) Cell {
        return self.ref(loc).*;
    }

    const CellIterator = struct {
        const CellAt = struct {
            data: Cell,
            at: Point,
        };

        origin: Point,
        width: usize,
        data: []Cell,
        i: usize = 0,
        cur: usize = 0,

        pub fn next(it: *@This()) ?Cell {
            if (it.i >= it.data.len) return null;
            it.cur = it.i;
            it.i += 1;
            return it.data[it.cur];
        }

        pub fn at(it: *@This()) Point {
            return it.origin + .{
                it.cur % it.width,
                @divTrunc(it.cur, it.width),
            };
        }

        pub fn relAt(it: *@This()) Size {
            return .{
                @intCast(u16, it.cur % it.width),
                @intCast(u16, @divTrunc(it.cur, it.width)),
            };
        }
    };

    pub fn cells(self: *Self) CellIterator {
        return .{
            .origin = .{
                self.bounds[0],
                self.bounds[1],
            },
            .width = self.width,
            .data = self.data,
        };
    }
};

const Grid = @import("grid.zig").Grid;

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

    var world = build: {
        var lines = Parse.lineScanner(input.reader());
        var builder = init: {
            var cur = try lines.next() orelse return error.NoInput;
            break :init Builder.initLine(arena.allocator(), cur) catch |err| return cur.carp(err);
        };
        var lineTime = try timing.timer(.parseLine);
        while (try lines.next()) |cur| {
            builder.parseLine(cur) catch |err| return cur.carp(err);
            try lineTime.lap();
        }
        break :build try builder.finish();
    };
    try timing.markPhase(.parse);

    var count: usize = 0;

    while (try world.pour()) count += 1;

    try timing.markPhase(.solve);

    try out.print(
        \\# Solution
        \\
    , .{});

    if (config.verbose > 0) {
        var size = rectSize(world.bounds);
        var grid = try Grid.init(allocator, .{
            .width = size[0],
            .height = size[1],
            .linePrefix = "    ",
        });
        defer grid.deinit();

        var worldCells = world.cells();
        while (worldCells.next()) |cell| {
            const gridAt = worldCells.relAt();
            grid.set(gridAt[0], gridAt[1], cell.glyph());
        }

        try out.print(
            \\@<{},{}>
            \\{}
            \\
        , .{
            world.bounds[0],
            world.bounds[1],
            grid,
        });
    }

    try out.print(
        \\> {}
        \\
    , .{count});

    try timing.markPhase(.report);

    try timing.finish(.overall);
}

const ArgParser = @import("args.zig").Parser;

const MainAllocator = std.heap.GeneralPurposeAllocator(.{
    // .verbose_log = true,
});

pub fn main() !void {
    var gpa = MainAllocator{};
    defer _ = gpa.deinit();

    var input = std.io.getStdIn();
    var output = std.io.getStdOut();
    var config = Config{};
    var bufferOutput = true;

    {
        var argsArena = std.heap.ArenaAllocator.init(gpa.allocator());
        defer argsArena.deinit();

        var args = try ArgParser.init(argsArena.allocator());
        defer args.deinit();

        // TODO: input filename arg

        while (args.next()) |arg| {
            if (arg.is(.{ "-h", "--help" })) {
                std.debug.print(
                    \\Usage: {s} [-v]
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
                config.verbose += 1;
            } else if (arg.is(.{"--raw-output"})) {
                bufferOutput = false;
            } else return error.InvalidArgument;
        }
    }

    var bufin = std.io.bufferedReader(input.reader());

    if (!bufferOutput)
        return run(gpa.allocator(), &bufin, output, config);

    var bufout = std.io.bufferedWriter(output.writer());
    try run(gpa.allocator(), &bufin, &bufout, config);
    try bufout.flush();
    // TODO: sentinel-buffered output writer to flush lines progressively
    // ... may obviate the desire for raw / non-buffered output else
}
