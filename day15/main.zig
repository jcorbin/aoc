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

        // \\               1    1    2    2
        // \\     0    5    0    5    0    5
        // \\ 0 ....S.......................
        // \\ 1 ......................S.....
        // \\ 2 ...............S............
        // \\ 3 ................SB..........
        // \\ 4 ............................
        // \\ 5 ............................
        // \\ 6 ............................
        // \\ 7 ..........S.......S.........
        // \\ 8 ............................
        // \\ 9 ............................
        // \\10 ....B.......................
        // \\11 ..S.........................
        // \\12 ............................
        // \\13 ............................
        // \\14 ..............S.......S.....
        // \\15 B...........................
        // \\16 ...........SB...............
        // \\17 ................S..........B
        // \\18 ....S.......................
        // \\19 ............................
        // \\20 ............S......S........
        // \\21 ............................
        // \\22 .......................B....

        \\ @{-2,0}
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

    const test_cases = [_]struct {
        input: []const u8,
        expected: []const u8,
        config: Config,
    }{

        // Part 1 example: setup
        .{
            .config = .{
                .verbose = 1,
            },
            .input = example_input,
            .expected = inital_state,
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
    allocator: Allocator,
    // TODO state to be built up line-to-line

    const Self = @This();

    pub fn initLine(allocator: Allocator, cur: *Parse.Cursor) !Self {
        var self = Self{
            .allocator = allocator,
        };
        try self.parseLine(cur);
        return self;
    }

    pub fn parseLine(self: *Self, cur: *Parse.Cursor) !void {
        _ = self;
        if (cur.live()) return error.ParseLineNotImplemented;
    }

    pub fn finish(self: *Self) !World {
        _ = self;
        return World{
            // TODO finalized problem data
        };
    }
};

const World = struct {
    // TODO problem representation
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
    try timing.markPhase(.parse);

    // FIXME: solve...
    _ = world;
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
