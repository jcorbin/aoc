const std = @import("std");

const mem = std.mem;
const Allocator = mem.Allocator;

test "example" {
    const example_input =
        \\Sensor at x=2, y=18: closest beacon is at x=-2, y=15
        \\Sensor at x=9, y=16: closest beacon is at x=10, y=16
        \\Sensor at x=13, y=2: closest beacon is at x=15, y=3
        \\Sensor at x=12, y=14: closest beacon is at x=10, y=16
        \\Sensor at x=10, y=20: closest beacon is at x=10, y=16
        \\Sensor at x=14, y=17: closest beacon is at x=10, y=16
        \\Sensor at x=8, y=7: closest beacon is at x=2, y=10
        \\Sensor at x=2, y=0: closest beacon is at x=2, y=10
        \\Sensor at x=0, y=11: closest beacon is at x=2, y=10
        \\Sensor at x=20, y=14: closest beacon is at x=25, y=17
        \\Sensor at x=17, y=20: closest beacon is at x=21, y=22
        \\Sensor at x=16, y=7: closest beacon is at x=15, y=3
        \\Sensor at x=14, y=3: closest beacon is at x=15, y=3
        \\Sensor at x=20, y=1: closest beacon is at x=15, y=3
        \\
    ;

    const inital_state =
        \\@{ -2, 0 }
        \\    ....S.......................
        \\    ......................S.....
        \\    ...............S............
        \\    ................SB..........
        \\    ............................
        \\    ............................
        \\    ............................
        \\    ..........S.......S.........
        \\    ............................
        \\    ............................
        \\    ....B.......................
        \\    ..S.........................
        \\    ............................
        \\    ............................
        \\    ..............S.......S.....
        \\    B...........................
        \\    ...........SB...............
        \\    ................S..........B
        \\    ....S.......................
        \\    ............................
        \\    ............S......S........
        \\    ............................
        \\    .......................B....
    ;

    //                1    1    2    2
    //      0    5    0    5    0    5
    // -2 ..........#.................
    // -1 .........###................
    //  0 ....S...#####...............
    //  1 .......#######........S.....
    //  2 ......#########S............
    //  3 .....###########SB..........
    //  4 ....#############...........
    //  5 ...###############..........
    //  6 ..#################.........
    //  7 .#########S#######S#........
    //  8 ..#################.........
    //  9 ...###############..........
    // 10 ....B############...........
    // 11 ..S..###########............
    // 12 ......#########.............
    // 13 .......#######..............
    // 14 ........#####.S.......S.....
    // 15 B........###................
    // 16 ..........#SB...............
    // 17 ................S..........B
    // 18 ....S.......................
    // 19 ............................
    // 20 ............S......S........
    // 21 ............................
    // 22 .......................B....

    const test_cases = [_]struct {
        input: []const u8,
        expected: []const u8,
        config: Config,
        skip: bool = false,
    }{

        // Part 1 example: setup
        .{
            .config = .{
                .verbose = 1,
            },
            .input = example_input,
            .expected = "" ++
                "# Init\n" ++
                inital_state ++
                "\n",
        },

        // Part 1 example: query y=10
        .{
            .config = .{
                .verbose = 1,
                .query_line = 10,
            },
            .input = example_input,
            .expected = 
            \\# Neighborhood around Y=10
            \\    ...#########################...
            \\    ..####B######################..
            \\    .###S#############.###########.
            \\
            \\> 26 eliminated cells
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
    query_line: ?i32 = null,
};

const Builder = struct {
    const Data = std.MultiArrayList(struct {
        sensor: Point,
        reading: Point,
    });

    allocator: Allocator,
    data: Data = .{},

    const Self = @This();

    pub fn initLine(allocator: Allocator, cur: *Parse.Cursor) !Self {
        var self = Self{
            .allocator = allocator,
        };
        try self.parseLine(cur);
        return self;
    }

    pub fn parseLine(self: *Self, cur: *Parse.Cursor) !void {
        cur.star(' ');
        if (!cur.haveLiteral("Sensor at")) return error.ExpectedSensorAt;
        const at = try self.parsePoint(cur);

        cur.star(' ');
        if (!cur.haveLiteral(":")) return error.ExpectedColon;

        cur.star(' ');
        if (!cur.haveLiteral("closest beacon is at")) return error.ExpectedClosestBeacon;
        const closest = try self.parsePoint(cur);

        if (cur.live()) return error.UnexpectedTrailer;

        try self.data.append(self.allocator, .{
            .sensor = at,
            .reading = closest,
        });
    }

    fn parsePoint(_: *Self, cur: *Parse.Cursor) !Point {
        cur.star(' ');
        if (!cur.haveLiteral("x=")) return error.ExpectedXEquals;
        const x = cur.consumeInt(i32, 10) orelse return error.ExpectedX;
        if (!cur.haveLiteral(",")) return error.ExpectedComma;
        cur.star(' ');
        if (!cur.haveLiteral("y=")) return error.ExpectedYEquals;
        const y = cur.consumeInt(i32, 10) orelse return error.ExpectedY;
        return Point{ x, y };
    }

    pub fn finish(self: *Self) !World {
        const slice = self.data.toOwnedSlice();

        return World{
            .allocator = self.allocator,
            .input_data = slice,

            .sensors = slice.items(.sensor),
            .readings = slice.items(.reading),
        };
    }
};

const space = @import("space.zig").Space(i32);
const Point = space.Point;

const World = struct {
    allocator: Allocator,
    input_data: Builder.Data.Slice,

    sensors: []const Point,
    readings: []const Point,

    const Self = @This();

    pub fn deinit(self: *Self) void {
        self.input_data.deinit(self.allocator);
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

    // FIXME: hookup your config
    _ = config;

    var arena = std.heap.ArenaAllocator.init(allocator);
    defer arena.deinit();

    var world = build: { // FIXME: parse input (store intermediate form, or evaluate)
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
    defer world.deinit();
    try timing.markPhase(.parse);

    if (config.verbose > 0) {
        var bounds = space.Rect{ .from = .{ 0, 0 }, .to = .{ 0, 0 } };
        for (world.sensors) |p| bounds.expandTo(p);
        for (world.readings) |p| bounds.expandTo(p);

        var grid = try Grid.init(arena.allocator(), .{
            .width = bounds.width(),
            .height = bounds.height(),
            .linePrefix = "    ",
            .fill = '.',
        });

        for (world.sensors) |p| {
            const r = bounds.relativize(p);
            grid.set(r[0], r[1], 'S');
        }

        for (world.readings) |p| {
            const r = bounds.relativize(p);
            grid.set(r[0], r[1], 'B');
        }

        try out.print(
            \\# Init
            \\@{}
            \\{}
            \\
        , .{ bounds.from, grid });
    }

    try timing.markPhase(.solve);

    if (config.query_line) |line| {
        try out.print(
            \\# Solution
            \\> {}
            \\
        , .{
            line,
        });
    }

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

        // TODO: input filename arg

        while (args.next()) |arg| {
            if (arg.is(.{ "-h", "--help" })) {
                std.debug.print(
                    \\Usage: {s} [-v]
                    \\
                    \\Options:
                    \\
                    \\  -q LINE or
                    \\  --query LINE
                    \\    query eliminated cells around a line
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
            } else if (arg.is(.{ "-q", "--query" })) {
                var line_arg = (try args.next()) orelse return error.MissingQueryLine;
                config.query_line = try line_arg.parseInt(i32, 10);
            } else if (arg.is(.{ "-v", "--verbose" })) {
                config.verbose += 1;
            } else if (arg.is(.{"--raw-output"})) {
                bufferOutput = false;
            } else return error.InvalidArgument;
        }
    }

    var bufin = std.io.bufferedReader(input.reader());

    if (!bufferOutput)
        return run(allocator, &bufin, output, config);

    var bufout = std.io.bufferedWriter(output.writer());
    try run(allocator, &bufin, &bufout, config);
    try bufout.flush();
    // TODO: sentinel-buffered output writer to flush lines progressively
    // ... may obviate the desire for raw / non-buffered output else
}
