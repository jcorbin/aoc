const std = @import("std");
const Allocator = std.mem.Allocator;
const Vector = std.meta.Vector;

test "example" {
    const test_cases = [_]struct {
        input: []const u8,
        expected: []const u8,
        config: Config,
    }{
        // Part 1 example
        .{
            .config = .{
                .verbose = true,
            },
            .input = 
            \\Sabqponm
            \\abcryxxl
            \\accszExk
            \\acctuvwj
            \\abdefghi
            \\
            ,
            .expected = 
            \\# Solution
            \\    v..v<<<<
            \\    >v.vv<<^
            \\    .>vv>E^^
            \\    ..v>>>^^
            \\    ..>>>>>^
            \\> 31 steps
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

const Grid = struct {
    allocator: Allocator,
    width: usize,
    lineOffset: usize,
    lineStride: usize,
    buf: []u8,

    const Self = @This();

    pub fn init(allocator: Allocator, opts: struct {
        width: usize,
        height: usize,
        linePrefix: []const u8 = "",
        lineSuffix: []const u8 = "\n",
        fill: u8 = ' ',
    }) !Self {
        const lineStride = opts.linePrefix.len + opts.width + opts.lineSuffix.len;
        const memSize = lineStride * opts.height;
        var buf = try allocator.alloc(u8, memSize);
        std.mem.set(u8, buf, opts.fill);

        if (opts.linePrefix.len > 0) {
            var i: usize = 0;
            while (i < buf.len) : (i += lineStride)
                std.mem.copy(u8, buf[i..], opts.linePrefix);
        }

        if (opts.lineSuffix.len > 0) {
            var i = opts.linePrefix.len + opts.width;
            while (i < buf.len) : (i += lineStride)
                std.mem.copy(u8, buf[i..], opts.lineSuffix);
        }

        return Self{
            .allocator = allocator,
            .width = opts.width,
            .lineOffset = opts.linePrefix.len,
            .lineStride = lineStride,
            .buf = buf,
        };
    }

    pub fn deinit(self: *Self) void {
        self.allocator.free(self.buf);
    }

    pub fn format(self: Self, comptime _: []const u8, _: std.fmt.FormatOptions, writer: anytype) !void {
        try writer.print("{s}", .{std.mem.trimRight(u8, self.buf, "\n")});
    }

    pub fn atPoint(self: Self, pt: Point) usize {
        return self.lineOffset + pointTo(usize, pt, self.lineStride);
    }

    pub fn ref(self: Self, pt: Point) *u8 {
        const i = self.atPoint(pt);
        return &self.buf[i];
    }

    pub fn set(self: Self, pt: Point, c: u8) void {
        self.ref(pt).* = c;
    }

    pub fn get(self: Self, pt: Point) u8 {
        return self.ref(pt).*;
    }
};

const Parse = @import("parse.zig");
const Timing = @import("perf.zig").Timing;

const Config = struct {
    verbose: bool = false,
};

const Cell = union(enum) {
    start: void,
    end: void,
    value: u5, // 26 possible heights

    const Self = @This();

    pub fn parse(c: u8) ?Self {
        return switch (c) {
            'S' => Self{ .start = {} },
            'E' => Self{ .end = {} },
            'a'...'z' => |l| Self{ .value = @intCast(u5, l - 'a') },
            else => null,
        };
    }

    pub fn height(self: Self) u5 {
        return switch (self) {
            .start => 0,
            .end => 25,
            .value => |n| n,
        };
    }
};

const Builder = struct {
    allocator: Allocator,
    buf: []Cell,
    width: u15,
    height: u15,
    at: u30 = 0,
    startAt: ?u30 = null,
    endAt: ?u30 = null,

    const Self = @This();

    pub fn initLine(allocator: Allocator, cur: *Parse.Cursor) !Self {
        const width = cur.buf.len;
        if (width > std.math.maxInt(u15)) return error.LineTooLong;
        var buf = try allocator.alloc(Cell, width * width);
        var self = Self{
            .allocator = allocator,
            .width = @intCast(u15, width),
            .height = 0,
            .buf = buf,
        };
        try self.parseLine(cur);
        return self;
    }

    pub fn parseLine(self: *Self, cur: *Parse.Cursor) !void {
        if (cur.buf.len != self.width) return error.WidthMismatch;
        if (self.height >= self.width) return error.HeightLimitExceede;
        while (self.at < self.buf.len) : (self.at += 1) {
            const b = cur.peek() orelse break;
            const c = Cell.parse(b) orelse return error.InvalidLetter;
            switch (c) {
                .start => {
                    if (self.startAt != null) return error.StartRedefined;
                    self.startAt = self.at;
                },
                .end => {
                    if (self.endAt != null) return error.EndRedefined;
                    self.endAt = self.at;
                },
                else => {},
            }
            self.buf[self.at] = c;
            cur.i += 1;
        }
        try cur.expectEnd(error.InputExceedsGrid);
        self.height += 1;
    }

    pub fn finish(self: *Self) !World {
        const startAt = self.startAt orelse return error.NoStart;
        const endAt = self.endAt orelse return error.NoEnd;
        const len = self.height * self.width;
        return .{
            .buf = self.buf[0..len],
            .width = self.width,
            .height = self.height,
            .startAt = startAt,
            .endAt = endAt,
        };
    }
};

const Point = Vector(2, i16);

fn pointFrom(at: u30, width: u15) Point {
    return .{
        @intCast(i16, at % width),
        @intCast(i16, at / width),
    };
}

fn pointTo(comptime T: anytype, pt: Point, width: anytype) T {
    return @intCast(T, pt[0]) + @intCast(T, pt[1]) * @intCast(T, width);
}

const World = struct {
    buf: []const Cell,
    width: u15,
    height: u15,
    startAt: u30,
    endAt: u30,

    const Self = @This();

    pub fn pointAt(self: Self, at: u30) Point {
        return pointFrom(at, self.width);
    }

    pub fn atPoint(self: Self, pt: Point) u30 {
        return pointTo(u30, pt, self.width);
    }

    pub fn get(self: Self, pt: Point) Cell {
        return self.buf[self.atPoint(pt)];
    }
};

const Solution = struct {
    const Move = enum {
        none,
        up,
        right,
        down,
        left,

        pub fn delta(self: @This()) Point {
            return switch (self) {
                .none => .{ 0, 0 },
                .up => .{ 0, -1 },
                .right => .{ 1, 0 },
                .down => .{ 0, 1 },
                .left => .{ -1, 0 },
            };
        }
    };

    const Data = std.MultiArrayList(struct {
        loc: Point,
        move: Move = .none,
    });

    allocator: Allocator,
    data: Data = .{},
    done: bool = false,

    // TODO any worth-it to having a hash set for loc?

    const Self = @This();
    pub fn init(allocator: Allocator, start: Point) !Self {
        var self = Self{ .allocator = allocator };
        self.data.append(self.allocator, .{ .loc = start });
        return self;
    }

    pub fn deinit(self: Self) void {
        self.data.deinit(self.allocator);
    }

    pub fn loc(self: Self) Point {
        return self.data.items(.loc)[self.data.len - 1];
    }

    pub fn then(self: Self, move: Move, to: Point, done: bool) !Self {
        var next = Self{
            .allocator = self.allocator,
            .data = try self.data.clone(self.allocator),
            .done = done,
        };
        errdefer next.data.deinit();
        next.data.items(.move)[next.data.len - 1] = move;
        try next.data.append(next.allocator, .{ .loc = to });
        return next;
    }

    const Next = struct {
        world: *const World,
        sol: Solution,
        loc: Point,
        i: usize = 0,

        const choices = [_]Move{ .up, .right, .down, .left };

        pub fn next(it: *@This()) !?Solution {
            if (!it.sol.done) {
                while (it.i < choices.len) {
                    const i = it.i;
                    it.i += 1;
                    if (try it.may(choices[i])) |sol| return sol;
                }
            }
            return null;
        }

        pub fn may(it: *@This(), move: Move) !?Solution {
            const to = it.loc + switch (move) {
                .none => return null,
                else => move.delta(),
            };
            if (to[0] < 0 or
                to[1] < 0 or
                to[0] >= it.world.width or
                to[1] >= it.world.height) return null;
            for (it.sol.data.items(.loc)) |prior|
                if (to == prior) return null;
            const done = it.world.atPoint(to) == it.world.endAt;
            return try it.sol.then(move, to, done);
        }
    };

    pub fn moves(self: Self, world: *const World) Next {
        return .{
            .world = world,
            .sol = self,
            .loc = self.loc(),
        };
    }
};

fn run(
    allocator: Allocator,

    // TODO: better "any .reader()-able / any .writer()-able" interfacing
    input: anytype,
    output: anytype,
    config: Config,
) !void {
    var timing = try Timing(enum {
        parse,
        parseLine,
        parseLineVerbose,
        solve,
        report,
        overall,
    }).start(allocator);
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
        while (try lines.next()) |cur| {
            var lineTime = try std.time.Timer.start();
            builder.parseLine(cur) catch |err| return cur.carp(err);
            try timing.collect(.parseLine, lineTime.lap());
        }
        break :build try builder.finish();
    };
    try timing.markPhase(.parse);

    { // FIXME: solve...
        std.debug.print("!!! TODO solve {} -> {}\n", .{ world.startAt, world.endAt });
    }
    try timing.markPhase(.solve);

    {
        var plan = try Grid.init(allocator, .{
            .width = world.width,
            .height = world.height,
            .linePrefix = "    ",
            .fill = '.',
        });

        plan.set(world.pointAt(world.startAt), 'S');
        plan.set(world.pointAt(world.endAt), 'E');

        try out.print(
            \\# Solution
            \\{}
            \\> {} steps
            \\
        , .{ plan, null });
    }
    try timing.markPhase(.report);

    try timing.finish(.overall);
}

const ArgParser = @import("args.zig").Parser;

pub fn main() !void {
    const allocator = std.heap.page_allocator;

    var input = std.io.getStdIn();
    var output = std.io.getStdOut();
    var config = Config{};
    var bufferOutput = true;

    {
        var args = try ArgParser.init(allocator);
        defer args.deinit();

        // TODO: input filename arg

        while (try args.next()) |arg| {
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
                config.trace = true;
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
