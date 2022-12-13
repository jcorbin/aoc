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
            // TODO this was the prompt example solution, but our search move
            //      biases don't match yet we find an equivalent path...
            //
            // \\    v..v<<<<
            // \\    >v.vv<<^
            // \\    .>vv>E^^
            // \\    ..v>>>^^
            // \\    ..>>>>>^
            //
            // \\    >>vv<<<<
            // \\    ..vvv<<^
            // \\    ..vv>E^^
            // \\    ..v>>>^^
            // \\    ..>>>>>^
            //
            \\    v..v<<<<
            \\    >v.vv<<^
            \\    .v.v>E^^
            \\    .>v>>>^^
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
const Timing = @import("perf.zig").Timing(enum {
    parse,
    parseLine,
    parseLineVerbose,
    search,
    searchRound,
    report,
    overall,
});

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

        var cost = try self.allocator.alloc(usize, len);
        std.mem.set(usize, cost, 0);

        return .{
            .width = self.width,
            .height = self.height,
            .startAt = startAt,
            .endAt = endAt,
            .cell = self.buf[0..len],
            .cost = cost,
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

fn pointSumSq(pt: Point) u32 {
    const x = @intCast(i32, pt[0]);
    const y = @intCast(i32, pt[1]);
    return @intCast(u32, x * x) + @intCast(u32, y * y);
}

const World = struct {
    width: u15,
    height: u15,
    startAt: u30,
    endAt: u30,
    cell: []const Cell,
    cost: []usize,

    const Self = @This();

    pub fn pointAt(self: Self, at: u30) Point {
        return pointFrom(at, self.width);
    }

    pub fn atPoint(self: Self, pt: Point) u30 {
        return pointTo(u30, pt, self.width);
    }

    pub fn get(self: Self, pt: Point) Cell {
        return self.cell[self.atPoint(pt)];
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

        pub fn glyph(self: @This()) u8 {
            return switch (self) {
                .none => '?',
                .up => '^',
                .right => '>',
                .down => 'v',
                .left => '<',
            };
        }
    };

    const Path = std.ArrayListUnmanaged(struct {
        loc: Point,
        move: Move = .none,
    });

    const Visited = std.DynamicBitSetUnmanaged;

    allocator: Allocator,
    done: bool = false,
    width: u15,
    height: u15,
    visited: Visited,
    path: Path,

    const Self = @This();

    const Search = struct {
        const Queue = std.PriorityQueue(*Self, *World, Self.compare);

        q: Queue,
        world: *World,
        best: ?*Self = null,

        pub fn run(allocator: Allocator, world: *World, timer: ?*Timing.Timer) !?*Self {
            var search = try Solution.Search.init(allocator, world);
            defer search.deinit();

            if (timer) |tm| tm.reset();
            while (search.q.removeOrNull()) |sol| {
                defer sol.destroy();
                try search.expand(sol);
                if (timer) |tm| try tm.lap();
            }

            var res = search.best;
            search.best = null;
            return res;
        }

        pub fn init(allocator: Allocator, world: *World) !Search {
            var q = Queue.init(allocator, world);

            var start = try Self.createStart(allocator, world);
            errdefer start.destroy();

            try q.add(start);
            return Search{ .q = q, .world = world };
        }

        pub fn deinit(search: *Search) void {
            if (search.best) |sol| sol.destroy();
            for (search.q.items[0..search.q.len]) |item| item.destroy();
            search.q.deinit();
        }

        fn isObsolete(search: *Search, sol: *const Self) bool {
            const prior = search.best orelse return false;
            return prior.done and prior.path.items.len <= sol.path.items.len;
        }

        fn isBetter(search: *Search, sol: *const Self) bool {
            if (!sol.done) return false;
            const best = search.best orelse return true;
            return sol.path.items.len < best.path.items.len;
        }

        fn expand(search: *Search, sol: *Self) !void {
            if (search.isObsolete(sol)) return;
            var solMoves = sol.moves(search.world);
            while (try solMoves.next()) |next| {
                if (search.isBetter(next)) {
                    var prior = search.best;
                    search.best = next;
                    if (prior) |s| s.destroy();
                } else if (next.done or search.isObsolete(next)) {
                    next.destroy();
                } else {
                    errdefer next.destroy();
                    try search.q.add(next);
                }
            }
        }
    };

    pub fn compare(
        world: *World,
        a: *const Self,
        b: *const Self,
    ) std.math.Order {
        if (a.done and !b.done) return .lt;
        if (b.done and !a.done) return .gt;

        const cost_cmp = std.math.order(a.path.items.len, b.path.items.len);

        const end_loc = world.pointAt(world.endAt);
        const dist_cmp = std.math.order(
            pointSumSq(end_loc - a.loc()),
            pointSumSq(end_loc - b.loc()),
        );

        // return switch (dist_cmp) {
        //     .eq => cost_cmp,
        //     else => |c| c.invert(),
        // };

        return switch (cost_cmp) {
            .eq => dist_cmp.invert(),
            else => |c| c,
        };
    }

    pub fn createStart(allocator: Allocator, world: *World) !*Self {
        return Self.create(
            allocator,
            world,
            world.pointAt(world.startAt),
        );
    }

    pub fn create(allocator: Allocator, world: *World, start: Point) !*Self {
        const width = world.width;
        const height = world.height;
        const len = width * height;

        // TODO: can we just make one contiguous allocation here and cast
        // tranches out of it?

        var self = try allocator.create(Self);
        errdefer allocator.destroy(self);

        var visited = try Visited.initEmpty(allocator, len);
        errdefer visited.deinit(allocator);

        var path = try Path.initCapacity(allocator, len);
        errdefer path.deinit(allocator);

        const startAt = pointTo(u30, start, width);
        visited.set(startAt);

        path.appendAssumeCapacity(.{ .loc = start });

        self.* = .{
            .allocator = allocator,
            .width = width,
            .height = height,
            .visited = visited,
            .path = path,
        };
        return self;
    }

    pub fn destroy(self: *Self) void {
        self.path.deinit(self.allocator);
        self.visited.deinit(self.allocator);
        self.allocator.destroy(self);
    }

    pub fn clone(self: *Self) !*Self {
        var copy = try self.allocator.create(Self);
        errdefer self.allocator.destroy(copy);

        var visited = try self.visited.clone(self.allocator);
        errdefer visited.deinit(self.allocator);

        var path = try self.path.clone(self.allocator);
        errdefer path.deinit(self.allocator);

        copy.* = .{
            .allocator = self.allocator,
            .done = self.done,
            .width = self.width,
            .height = self.height,
            .visited = visited,
            .path = path,
        };
        return copy;
    }

    pub fn loc(self: *const Self) Point {
        return self.path.items[self.path.items.len - 1].loc;
    }

    pub fn final(self: *const Self) Move {
        return self.path.items[self.path.items.len - 1].move;
    }

    const ThenError = std.mem.Allocator.Error || error{
        SolutionInvalid,
    };

    pub fn may(self: *Self, world: *World, move: Move) ThenError!?*Self {
        if (self.done) return null;

        // path was allocated with enough capacity to visit every cell in width x height
        // so if we're out of capacity, this solution should be abandoned
        const pathLen = self.path.items.len;
        if (pathLen >= self.path.capacity)
            return ThenError.SolutionInvalid;
        const last = pathLen - 1;

        const from = self.loc();
        const to = from + switch (move) {
            .none => return null,
            else => move.delta(),
        };

        // bounds check
        if (to[0] < 0 or
            to[1] < 0 or
            to[0] >= self.width or
            to[1] >= self.height) return null;
        const toAt = pointTo(u30, to, self.width);

        // loop check
        if (self.visited.isSet(toAt)) return null;

        // climbing check
        const from_level = world.get(from).height();
        const to_level = world.get(to).height();
        if (from_level < to_level and
            to_level - from_level > 1)
            return null;

        // cost check / update
        const prior = world.cost[toAt];
        if (prior > 0 and prior <= pathLen) return null;
        world.cost[toAt] = pathLen;

        var next = try self.clone();
        next.done = toAt == world.endAt;
        next.visited.set(toAt);
        next.path.items[last].move = move;
        next.path.appendAssumeCapacity(.{ .loc = to });
        return next;
    }

    const Next = struct {
        world: *World,
        sol: *Solution,
        loc: Point,
        i: usize = 0,

        const choices = [_]Move{
            .up,
            .down,
            .left,
            .right,
        };

        pub fn next(it: *@This()) !?*Solution {
            if (it.sol.done) return null;
            while (it.i < choices.len) {
                const i = it.i;
                it.i += 1;
                if (try it.sol.may(it.world, choices[i])) |sol| return sol;
            }
            return null;
        }
    };

    pub fn moves(self: *Self, world: *World) Next {
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

    var searchRoundTimer = try timing.timer(.searchRound);
    var solution = try Solution.Search.run(
        allocator,
        &world,
        &searchRoundTimer,
    ) orelse return error.NoSolution;
    defer solution.destroy();
    try timing.markPhase(.search);

    {
        var plan = try Grid.init(allocator, .{
            .width = world.width,
            .height = world.height,
            .linePrefix = "    ",
            .fill = '.',
        });
        defer plan.deinit();

        plan.set(world.pointAt(world.startAt), 'S');
        plan.set(world.pointAt(world.endAt), 'E');

        { // TODO maybe into a Solution.method()
            for (solution.path.items) |item|
                switch (item.move) {
                    .none => {},
                    else => |move| plan.set(item.loc, move.glyph()),
                };
        }

        var steps = solution.path.items.len;
        if (solution.final() == .none) steps -= 1;
        try out.print(
            \\# Solution
            \\{}
            \\> {} steps
            \\
        , .{ plan, steps });
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
                config.verbose = true;
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
