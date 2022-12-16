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

    const test_cases = [_]struct {
        input: []const u8,
        expected: []const u8,
        config: Config,
        skip: bool = false,
    }{

        // Part 1 example: setup
        .{
            .config = .{
                .verbose = 2,
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
                .query = .{ .line = 10 },
            },
            .input = example_input,
            .expected = 
            \\# Neighborhood around Y=10
            \\    .#########################..
            \\    ####B######################.
            \\    ##S#############.###########
            \\> 26 eliminated cells
            \\
            ,
        },

        // Part 2 example: tune 20
        .{
            .config = .{
                .verbose = 1,
                .query = .{
                    .tune = .{ .upto = 20, .by = 4000000 },
                },
            },
            .input = example_input,
            .expected = 
            \\# Tune to 20
            \\@{ 14, 11 }
            \\> 56000011
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
    query: union(enum) {
        none: void,
        line: i32,
        tune: struct {
            upto: u32,
            by: u32,
        },
    } = .none,
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

    pub fn finish(self: *Self) World {
        const slice = self.data.toOwnedSlice();

        return World{
            .allocator = self.allocator,
            .input_data = slice,

            .sensors = slice.items(.sensor),
            .readings = slice.items(.reading),

            .query = RangeList.init(self.allocator),
        };
    }
};

const space = @import("space.zig").Space(i32);
const Point = space.Point;

fn dist(a: Point, b: Point) usize {
    const dx = if (a[0] > b[0]) a[0] - b[0] else b[0] - a[0];
    const dy = if (a[1] > b[1]) a[1] - b[1] else b[1] - a[1];
    return @intCast(usize, dx) + @intCast(usize, dy);
}

const World = struct {
    allocator: Allocator,
    input_data: Builder.Data.Slice,

    sensors: []const Point,
    readings: []const Point,
    areas: []const Area = &[_]Area{},
    query: RangeList,

    const Self = @This();

    pub fn deinit(self: *Self) void {
        self.input_data.deinit(self.allocator);
    }

    pub fn bounds(self: Self) space.Rect {
        var r = space.Rect{ .from = .{ 0, 0 }, .to = .{ 0, 0 } };
        for (self.sensors) |p| r.expandTo(p);
        for (self.readings) |p| r.expandTo(p);
        return r;
    }

    pub fn collect_areas(self: *Self) !void {
        var areas = try self.allocator.alloc(Area, self.sensors.len);
        for (self.sensors) |at, i| {
            const reading = self.readings[i];
            const r = dist(at, reading);
            areas[i] = .{
                .at = at,
                .r = r,
                .lo = at - @splat(2, @intCast(i32, r)),
                .hi = at + @splat(2, @intCast(i32, r)),
            };
        }
        self.areas = areas;
    }

    pub fn query_at(self: *Self, y: i32) !void {
        self.query.data.clearRetainingCapacity();
        try self.query.data.ensureTotalCapacity(self.areas.len * 2);
        for (self.areas) |area|
            if (area.x_range(y)) |range|
                try self.query.addRange(range);
    }

    pub fn range_size(self: *Self) usize {
        var count: usize = 0;
        for (self.query.data.items) |range| count += range.size();
        return count;
    }
};

const Area = struct {
    at: Point,
    r: usize,
    lo: Point,
    hi: Point,

    pub fn x_range(area: @This(), y: i32) ?Range {
        if (area.lo[1] <= y and y <= area.hi[1]) {
            const d = if (area.at[1] > y) area.at[1] - y else y - area.at[1];
            const r = @intCast(i32, area.r) - d;
            return .{
                .start = area.at[0] - r,
                .end = area.at[0] + r + 1,
            };
        } else return null;
    }
};

const Grid = @import("grid.zig").Grid;

const Range = struct {
    start: i32,
    end: i32,

    const Self = @This();

    pub fn point(at: i32) Self {
        return .{ .start = at, .end = at + 1 };
    }

    pub fn size(self: Self) usize {
        return if (self.start < self.end) @intCast(usize, self.end - self.start) else 0;
    }

    pub fn includes(self: Self, x: i32) bool {
        return self.start <= x and x < self.end;
    }

    pub fn valid(self: Self) bool {
        return self.start < self.end;
    }

    pub fn merge(a: Self, b: Self) ?Self {
        if (!(a.valid() and b.valid())) return null;
        if (a.start > b.start) return merge(b, a);
        if (a.end <= b.start) return null;
        return .{
            .start = @minimum(a.start, b.start),
            .end = @maximum(a.end, b.end),
        };
    }

    pub fn startLessThan(_: void, a: Self, b: Self) bool {
        return a.start < b.start;
    }
};

pub fn search(
    comptime T: type,
    key: T,
    items: []const T,
    context: anytype,
    comptime where: fn (context: @TypeOf(context), lhs: T, rhs: T) bool,
) usize {
    var left: usize = 0;
    var right: usize = items.len;
    while (left < right) {
        const mid = left + (right - left) / 2;
        if (where(context, key, items[mid])) {
            right = mid;
        } else {
            left = mid + 1;
        }
    }
    return left;
}

const RangeList = struct {
    const Data = std.ArrayList(Range);

    data: Data,

    const Self = @This();

    pub fn init(allocator: Allocator) Self {
        return Self{
            .data = Data.init(allocator),
        };
    }

    pub fn addRange(self: *Self, range: Range) !void {
        const i = search(Range, range, self.data.items, {}, Range.startLessThan);
        if (i > 0 and try self.mayMerge(range, i - 1))
            return
        else if (try self.mayMerge(range, i))
            return
        else
            try self.data.insert(i, range);
    }

    fn mayMerge(self: *Self, range: Range, i: usize) !bool {
        if (i >= self.data.items.len) return false;
        var merged = self.data.items[i].merge(range) orelse return false;
        try self.data.replaceRange(i, 1, &[_]Range{merged});

        const j = i + 1;
        while (j < self.data.items.len) {
            merged = merged.merge(self.data.items[j]) orelse break;
            try self.data.replaceRange(i, 2, &[_]Range{merged});
        }

        return true;
    }

    pub fn removePoint(self: *Self, x: i32) !void {
        for (self.data.items) |range, j| {
            if (range.includes(x)) {
                const a = Range{ .start = range.start, .end = x };
                const b = Range{ .start = x + 1, .end = range.end };
                try self.data.replaceRange(j, 1, if (a.valid() and b.valid())
                    &[_]Range{ a, b }
                else if (a.valid())
                    &[_]Range{a}
                else if (b.valid())
                    &[_]Range{b}
                else
                    &[_]Range{});
                break;
            }
        }
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
        break :build builder.finish();
    };
    defer world.deinit();
    try timing.markPhase(.parse);

    const bounds = world.bounds();
    const size = bounds.size();

    if (config.verbose > 1) {
        if (size[0] > 1_000 or size[1] > 1_000) {
            try out.print(
                \\# Init (Large!)
                \\@{} size: {}
                \\
            , .{ bounds.from, size });
        } else {
            var grid = try Grid.init(arena.allocator(), .{
                .width = size[0],
                .height = size[1],
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
    }

    try world.collect_areas();

    if (config.verbose > 1) {
        if (size[0] > 1_000 or size[1] > 1_000) {
            try out.print(
                \\# Areas (Large!)
                \\@{} size: {}
                \\
            , .{ bounds.from, size });
        } else {
            var grid = try Grid.init(arena.allocator(), .{
                .width = size[0],
                .height = size[1],
                .linePrefix = "        ",
                .fill = '.',
            });

            {
                var y = bounds.from[1];
                var offset: usize = 4;
                while (y < bounds.to[1]) : (y += 1) {
                    _ = try std.fmt.bufPrint(grid.buf[offset .. offset + 4], "{d: <4}", .{y});
                    offset += grid.lineStride;
                }
            }

            for (world.areas) |area| {
                var y = bounds.from[1];

                while (y < bounds.to[1]) : (y += 1) {
                    if (area.x_range(y)) |range| {
                        var x = range.start;
                        while (x < range.end) : (x += 1) {
                            const p = Point{ x, y };
                            if (bounds.contains(p)) {
                                const r = bounds.relativize(p);
                                grid.set(r[0], r[1], '#');
                            }
                        }
                    }
                }
            }

            for (world.sensors) |p| {
                const r = bounds.relativize(p);
                grid.set(r[0], r[1], 'S');
            }

            for (world.readings) |p| {
                const r = bounds.relativize(p);
                grid.set(r[0], r[1], 'B');
            }

            try out.print(
                \\# Areas
                \\@{}
                \\{}
                \\
            , .{ bounds.from, grid });
        }
    }

    switch (config.query) {
        .none => {},
        .line => |line| {
            try out.print("# Neighborhood around Y={}\n", .{line});

            try world.query_at(line);
            inline for ([_][]const Point{ world.sensors, world.readings }) |points| {
                for (points) |at| if (at[1] == line)
                    try world.query.removePoint(at[0]);
            }

            const count = world.range_size();
            try timing.markPhase(.solve);

            if (config.verbose > 0) {
                const show_lines = [_]i32{ line - 1, line, line + 1 };

                const show_bounds = space.Rect{
                    .from = space.Point{ bounds.from[0], show_lines[0] },
                    .to = space.Point{ bounds.to[0], show_lines[2] + 1 },
                };

                var grid = try Grid.init(arena.allocator(), .{
                    .width = size[0], // TODO truncate to affected
                    .height = show_lines.len,
                    .linePrefix = "    ",
                    .fill = '.',
                });

                for (world.areas) |area| {
                    for (show_lines) |show_y| {
                        if (area.x_range(show_y)) |range| {
                            const y = @intCast(usize, show_y - show_bounds.from[1]);

                            var x = @intCast(usize, @maximum(0, range.start - show_bounds.from[0]));
                            const until = @minimum(grid.width, @intCast(usize, range.end - show_bounds.from[0]));
                            while (x < until) : (x += 1) grid.set(x, y, '#');
                        }
                    }
                }

                inline for ([_]struct { where: []const Point, mark: u8 }{
                    .{ .where = world.sensors, .mark = 'S' },
                    .{ .where = world.readings, .mark = 'B' },
                }) |wm| for (wm.where) |at| if (show_bounds.contains(at)) {
                    const p = show_bounds.relativize(at);
                    grid.set(p[0], p[1], wm.mark);
                };

                try out.print("{}\n", .{grid});
            }

            try out.print("> {} eliminated cells\n", .{count});

            try timing.markPhase(.report);
        },
        .tune => |tune| {
            var y: i32 = 0;
            while (y <= tune.upto) : (y += 1) {
                // TODO debug extra range holes, the areas dump above is
                // correct, but query range lists below have too many gap
                //
                //                 1    1    2
                //       0    5    0    5    0
                // 0  ## ##S################## #####
                // 1  ## ####################S #####
                // 2  ## #############S####### #####
                // 3  ## ##############SB##### #####
                // 4  ## ##################### ####.
                // 5  ## ##################### ###..
                // 6  ## ##################### ##...
                // 7  .# ########S#######S#### #....
                // 8  .. ##################### ##...
                // 9  .# ##################### ###..
                // 10 ## ##B################## ####.
                // 11 ## S#############.###### #####
                // 12 ## ##################### #####
                // 13 .# ##################### #####
                // 14 .# ############S#######S #####
                // 15 B# ##################### #####
                // 16 ## #########SB########## #####
                // 17 ## ##############S###### ####B
                // 18 ## ##S################## ####.
                // 19 ## ##################### ###..
                // 20 ## ##########S######S### ###..
                //
                // 21 ## ##################### ##...
                // 22 .# ######..############# B....

                try out.print("- query_at: {}\n", .{y});

                try world.query_at(y);
                inline for ([_][]const Point{ world.sensors, world.readings }) |points| {
                    for (points) |at| if (at[1] == y) {
                        const p = Range.point(at[0]);
                        try out.print("  - add point: {}\n", .{p});
                        try world.query.addRange(p);
                    };
                }

                for (world.query.data.items) |r, i|
                    try out.print("  {}. {}\n", .{ i, r });
            }
        },
    }

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
            } else if (arg.is(.{ "-t", "--tune" })) {
                var line_arg = args.next() orelse return error.MissingQueryLine;
                config.query = .{ .tune = .{
                    .upto = try line_arg.parseInt(u32, 10),
                    .by = 4000000,
                } };
            } else if (arg.is(.{ "-q", "--query" })) {
                var line_arg = args.next() orelse return error.MissingQueryLine;
                config.query = .{ .line = try line_arg.parseInt(i32, 10) };
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
