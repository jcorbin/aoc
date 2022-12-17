const std = @import("std");

const assert = std.debug.assert;
const mem = std.mem;

const Allocator = mem.Allocator;

test "example" {
    const example_input =
        \\Valve AA has flow rate=0; tunnels lead to valves DD, II, BB
        \\Valve BB has flow rate=13; tunnels lead to valves CC, AA
        \\Valve CC has flow rate=2; tunnels lead to valves DD, BB
        \\Valve DD has flow rate=20; tunnels lead to valves CC, AA, EE
        \\Valve EE has flow rate=3; tunnels lead to valves FF, DD
        \\Valve FF has flow rate=0; tunnels lead to valves EE, GG
        \\Valve GG has flow rate=0; tunnels lead to valves FF, HH
        \\Valve HH has flow rate=22; tunnel leads to valve GG
        \\Valve II has flow rate=0; tunnels lead to valves AA, JJ
        \\Valve JJ has flow rate=21; tunnel leads to valve II
        \\
    ;

    const test_cases = [_]struct {
        input: []const u8,
        expected: []const u8,
        config: Config,
        skip: bool = false,
    }{
        // Part 1 example
        .{
            .config = .{
                .verbose = 1,
            },
            .input = example_input,
            .expected = 
            \\# Minute 1
            \\- No valves are open.
            \\- You move to valve DD.
            \\
            \\# Minute 2
            \\- No valves are open.
            \\- You open valve DD.
            \\
            \\# Minute 3
            \\- Valve DD is open, releasing 20 pressure.
            \\- You move to valve CC.
            \\
            \\# Minute 4
            \\- Valve DD is open, releasing 20 pressure.
            \\- You move to valve BB.
            \\
            \\# Minute 5
            \\- Valve DD is open, releasing 20 pressure.
            \\- You open valve BB.
            \\
            \\# Minute 6
            \\- Valves BB and DD are open, releasing 33 pressure.
            \\- You move to valve AA.
            \\
            \\# Minute 7
            \\- Valves BB and DD are open, releasing 33 pressure.
            \\- You move to valve II.
            \\
            \\# Minute 8
            \\- Valves BB and DD are open, releasing 33 pressure.
            \\- You move to valve JJ.
            \\
            \\# Minute 9
            \\- Valves BB and DD are open, releasing 33 pressure.
            \\- You open valve JJ.
            \\
            \\# Minute 10
            \\- Valves BB, DD, and JJ are open, releasing 54 pressure.
            \\- You move to valve II.
            \\
            \\# Minute 11
            \\- Valves BB, DD, and JJ are open, releasing 54 pressure.
            \\- You move to valve AA.
            \\
            \\# Minute 12
            \\- Valves BB, DD, and JJ are open, releasing 54 pressure.
            \\- You move to valve DD.
            \\
            \\# Minute 13
            \\- Valves BB, DD, and JJ are open, releasing 54 pressure.
            \\- You move to valve EE.
            \\
            \\# Minute 14
            \\- Valves BB, DD, and JJ are open, releasing 54 pressure.
            \\- You move to valve FF.
            \\
            \\# Minute 15
            \\- Valves BB, DD, and JJ are open, releasing 54 pressure.
            \\- You move to valve GG.
            \\
            \\# Minute 16
            \\- Valves BB, DD, and JJ are open, releasing 54 pressure.
            \\- You move to valve HH.
            \\
            \\# Minute 17
            \\- Valves BB, DD, and JJ are open, releasing 54 pressure.
            \\- You open valve HH.
            \\
            \\# Minute 18
            \\- Valves BB, DD, HH, and JJ are open, releasing 76 pressure.
            \\- You move to valve GG.
            \\
            \\# Minute 19
            \\- Valves BB, DD, HH, and JJ are open, releasing 76 pressure.
            \\- You move to valve FF.
            \\
            \\# Minute 20
            \\- Valves BB, DD, HH, and JJ are open, releasing 76 pressure.
            \\- You move to valve EE.
            \\
            \\# Minute 21
            \\- Valves BB, DD, HH, and JJ are open, releasing 76 pressure.
            \\- You open valve EE.
            \\
            \\# Minute 22
            \\- Valves BB, DD, EE, HH, and JJ are open, releasing 79 pressure.
            \\- You move to valve DD.
            \\
            \\# Minute 23
            \\- Valves BB, DD, EE, HH, and JJ are open, releasing 79 pressure.
            \\- You move to valve CC.
            \\
            \\# Minute 24
            \\- Valves BB, DD, EE, HH, and JJ are open, releasing 79 pressure.
            \\- You open valve CC.
            \\
            \\# Minute 25
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 26
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 27
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 28
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 29
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 30
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Solution
            \\> 1651 total pressure released
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

fn memAsc(comptime T: type) fn (void, T, T) bool {
    const impl = struct {
        fn inner(_: void, a: []const T, b: []const T) bool {
            return std.mem.lessThan(T, a, b);
        }
    };
    return impl.inner;
}

const Parse = @import("parse.zig");
const Timing = @import("perf.zig").Timing(enum {
    parse,
    parseLine,

    searchRound,
    solve,

    report,
    overall,
});

const Config = struct {
    verbose: usize = 0,
};

const Builder = struct {
    const Data = std.ArrayListUnmanaged(union(enum) {
        name: []const u8,
        flow: u16,
        next: []const u8,
    });

    allocator: Allocator,
    arena: std.heap.ArenaAllocator,
    data: Data = .{},

    const Self = @This();

    pub fn initLine(allocator: Allocator, cur: *Parse.Cursor) !Self {
        var self = Self{
            .allocator = allocator,
            .arena = std.heap.ArenaAllocator.init(allocator),
        };
        try self.parseLine(cur);
        return self;
    }

    pub fn intern(self: *Self, str: []const u8) ![]const u8 {
        // TODO: a self.strings : std.StringHashMap
        return try self.arena.allocator().dupe(u8, str);
    }

    pub fn parseLine(self: *Self, cur: *Parse.Cursor) !void {
        try self.data.ensureUnusedCapacity(self.allocator, 3);

        _ = cur.star(.{ .just = ' ' });
        _ = cur.have(.{ .literal = "Valve" }) orelse return error.ExpectedValve;

        _ = cur.star(.{ .just = ' ' });
        self.data.appendAssumeCapacity(.{
            .name = try self.intern(cur.haveN(2, .{
                .range = .{ .min = 'A', .max = 'Z' },
            }) orelse return error.ExpectedValveName),
        });

        _ = cur.star(.{ .just = ' ' });
        _ = cur.have(.{ .literal = "has flow rate=" }) orelse return error.ExpectedFlowRate;
        self.data.appendAssumeCapacity(.{
            .flow = try std.fmt.parseInt(u16, cur.plus(.{
                .any = "0123456789",
            }) orelse return error.ExpectedNumber, 10),
        });

        if (cur.have(.{ .literal = "; tunnel leads to valve" }) != null) {
            _ = cur.star(.{ .just = ' ' });
            try self.data.append(self.allocator, .{
                .next = try self.intern(cur.haveN(2, .{
                    .range = .{ .min = 'A', .max = 'Z' },
                }) orelse return error.ExpectedValveName),
            });
            if (cur.live()) return error.UnexpectedTrailer;
        } else if (cur.have(.{ .literal = "; tunnels lead to valves" }) != null) {
            while (cur.live()) {
                _ = cur.star(.{ .just = ' ' });
                try self.data.append(self.allocator, .{
                    .next = try self.intern(cur.haveN(2, .{
                        .range = .{ .min = 'A', .max = 'Z' },
                    }) orelse return error.ExpectedValveName),
                });
                if (cur.live())
                    _ = cur.have(.{ .just = ',' }) orelse return error.ExpectedComma;
            }
        } else return error.ExpectedConnectedValves;
    }

    pub fn finish(self: *Self) !World {
        errdefer self.arena.deinit();
        defer self.data.deinit(self.allocator);

        // to cause next-linkage flush below
        try self.data.append(self.allocator, .{ .name = "" });

        var slab = try self.arena.allocator().alloc(Valve, count: {
            var n: usize = 0;
            for (self.data.items) |item| switch (item) {
                .name => |name| {
                    if (name.len > 0) n += 1;
                },
                else => {},
            };
            break :count n;
        });

        var links = try self.arena.allocator().alloc(*Valve, count: {
            var n: usize = 0;
            for (self.data.items) |item| switch (item) {
                .next => n += 1,
                else => {},
            };
            break :count n;
        });

        var valves = World.ValveMap{};
        try valves.ensureTotalCapacity(self.arena.allocator(), @intCast(u32, slab.len));

        for (self.data.items) |item| switch (item) {
            .name => |name| if (name.len > 0) {
                var valve = &slab[0];
                slab = slab[1..];
                valve.* = .{
                    .name = name,
                    .next = links[0..0],
                };
                try valves.put(self.arena.allocator(), name, valve);
            },
            else => {},
        };
        assert(slab.len == 0);

        var valve: ?*Valve = null;
        var next_len: usize = 0;
        for (self.data.items) |item| switch (item) {
            .name => |name| {
                if (valve) |prior| {
                    prior.next = links[0..next_len];
                    links = links[next_len..];
                    next_len = 0;
                }
                valve = if (name.len > 0)
                    valves.get(name) orelse return error.UndefinedValve
                else
                    null;
            },
            .flow => |flow| {
                if (valve) |prior| prior.flow = flow else return error.FlowBeforeValve;
            },
            .next => |name| {
                links[next_len] = valves.get(name) orelse return error.UndefinedNextValve;
                next_len += 1;
            },
        };
        assert(links.len == 0);

        return World{
            .allocator = self.allocator,
            .arena = self.arena,
            .valves = valves,
        };
    }
};

const Valve = struct {
    name: []const u8,
    flow: u16 = 0,
    next: []*Valve,

    const Self = @This();

    pub fn format(self: Self, comptime _: []const u8, _: std.fmt.FormatOptions, writer: anytype) !void {
        try writer.print("{s}({})", .{ self.name, self.flow });
        for (self.next) |next, i|
            if (i == 0)
                try writer.print(" -> {s}", .{next.name})
            else
                try writer.print(", {s}", .{next.name});
    }
};

const World = struct {
    const ValveMap = std.StringHashMapUnmanaged(*Valve);

    allocator: Allocator,
    arena: std.heap.ArenaAllocator,
    valves: ValveMap,

    const Self = @This();

    pub fn deinit(self: *Self) void {
        self.arena.deinit();
    }
};

const search = @import("search.zig");

const Search = struct {
    const Self = @This();

    const Queue = search.Queue(
        Plan,
        Expander,
        *Self,
        comparePlan,
        expandPlan,
        consumePlan,
        destroyPlan,
    );

    allocator: Allocator,
    queue: Queue,
    world: *const World,
    result: ?Plan,

    pub fn init(allocator: Allocator, opts: struct {
        world: *const World,
        max_steps: usize,
        start_name: []const u8,
    }) !*Self {
        var self = try allocator.create(Self);
        errdefer allocator.free(self);
        self.* = .{
            .allocator = allocator,
            .queue = try Queue.init(allocator, self),
            .world = opts.world,
        };

        try self.queue.add(Plan.init(
            allocator,
            self.world.valves.get(self.start_name) orelse return error.NoStartValve,
            self.max_steps,
        ));

        return self;
    }

    pub fn deinit(self: *Self) void {
        self.queue.deinit();
        self.allocator.free(self);
    }

    fn consumePlan(self: *Self, plan: Plan) search.Action {
        if (plan.steps.items.len < plan.steps.capacity)
            return .queue;
        if (self.result) |prior| {
            if (plan.totalReleased > prior.totalReleased) {
                self.result = plan;
                destroyPlan(prior);
            } else {
                destroyPlan(plan);
            }
        } else self.result = plan;
        return .take;
    }

    fn destroyPlan(self: *Self, plan: Plan) void {
        plan.deinit(self.allocator);
    }

    fn comparePlan(_: *Self, a: Plan, b: Plan) std.math.Order {
        // NOTE .lt -> a expandBefore b

        const stepOrder = std.math.order(a.availableSteps(), b.availableSteps());

        // deepen paths first to keep search frontier smaller
        return stepOrder.invert();

        // TODO totalReleased doesn't really matter yet, since we have no
        //      pruning and will need to explore every path eventually anyhow,
        //      but for if we did/can prune:
        // switch (stepOrder.invert()) {
        //     // explore better plans first
        //     .eq => return std.math.order(a.totalReleased, b.totalReleased).invert(),
        //     // deepen paths first to keep search frontier smaller
        //     else => |ord| return ord,
        // }
    }

    fn expandPlan(self: *Self, plan: Plan) Expander {
        return .{ .self = self, .prior = plan };
    }

    const Expander = struct {
        self: *Self,
        prior: Plan,
        opened: bool = false,
        nexti: usize = 0,
        reused: bool = false,

        pub fn next(it: *@This()) !?Plan {
            // TODO stop if no more can be opened

            if (!it.opened) {
                it.opened = true;
                if (it.prior.at.flow > 0 and
                    it.prior.opened.get(it.prior.at) == null)
                {
                    var plan = it.nextPlan();
                    try plan.open();
                    return plan;
                }
            }

            if (it.nextMove()) |move| {
                var plan = it.nextPlan();
                try plan.moveTo(move);
                return plan;
            }

            return null;
        }

        fn nextMove(it: *@This()) *const Valve {
            const nx = it.prior.at.nx;
            if (it.nexti < nx.len) {
                it.nexti += 1;
                return nx[it.nexti];
            }
            return null;
        }

        fn nextPlan(it: *@This()) !Plan {
            if (it.reused) return error.PlanReused;
            const nx = it.prior.at.nx;
            if (it.nexti < nx.len) {
                return try it.prior.clone(it.self.allocator);
            } else {
                it.reused = true;
                return it.prior;
            }
        }

        pub fn deinit(it: @This()) void {
            if (!it.reuse)
                it.prior.deinit(it.self.allocator);
        }
    };
};

const ValveSet = std.AutoHashMapUnmanaged(*const Valve, void);

const Plan = struct {
    const Step = union(enum) {
        move: *const Valve,
        open: *const Valve,
    };

    const Steps = std.ArrayListUnmanaged(Step);

    at: *Valve,
    steps: Steps,
    opened: ValveSet = .{},
    totalOpen: usize = 0,
    totalReleased: usize = 0,

    const Self = @This();

    pub fn init(allocator: Allocator, start: *Valve, max_steps: usize) !Self {
        var self = Self{
            .at = start,
            .steps = try Steps.initCapacity(allocator, max_steps),
        };
        try self.opened.ensureUnusedCapacity(max_steps);
        return self;
    }

    pub fn clone(self: Self, allocator: Allocator) !Self {
        var other = self;

        other.steps = try self.steps.clone(allocator);
        errdefer other.steps.deinit(allocator);

        other.opened = try self.opened.clone(allocator);
        errdefer other.opened.deinit(allocator);

        return other;
    }

    pub fn deinit(self: Self, allocator: Allocator) void {
        self.steps.deinit(allocator);
        self.opened.deinit(allocator);
    }

    pub fn moveTo(self: *Self, move: *const Valve) !void {
        if (self.steps.items.len >= self.steps.capacity) return error.PlanFull;
        self.at = move;
        self.steps.appendAssumeCapacity(.{ .move = move });
        self.totalReleased += self.totalOpen;
    }

    pub fn open(self: *Self) !void {
        if (self.steps.items.len >= self.steps.capacity) return error.PlanFull;
        const valve = self.at;
        assert(valve.flow > 0);
        self.steps.appendAssumeCapacity(.{ .open = valve });
        self.opened.putAssumeCapacity(valve, {});
        self.totalReleased += self.totalOpen;
        self.totalOpen += valve.flow;
    }

    pub fn availableSteps(self: Self) usize {
        return self.steps.capacity - self.steps.items.len;
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
    defer world.deinit();
    try timing.markPhase(.parse);

    var srch = try Search.init(arena.allocator(), .{
        .world = &world,
        .max_steps = 30,
        .start_name = "AA",
    });
    defer srch.deinit();
    while (!srch.done()) {
        var roundTime = try timing.timer(.searchRound);
        const ran = srch.runUpto(1_000);
        try roundTime.lap();

        if (config.verbose > 0) {
            const time = timing.data.items[timing.data.len - 1].time;
            const best = if (srch.result) |res| res.totalReleased else 0;
            std.debug.print("searched {} in {} ; best = {}\n", .{ ran, time, best });
        }
    }
    const result = srch.result orelse return error.NoResultFound;
    try timing.markPhase(.solve);

    if (config.verbose > 0) {
        var opened = ValveSet{};
        try opened.ensureUnusedCapacity(arena.allocator(), result.steps.len);
        var totalOpen: usize = 0;
        var nameList = try arena.allocator().alloc([]const u8, opened.capacity);
        var tmp = std.ArrayList(u8).init(arena.allocator());

        for (result.steps) |step, step_i| {
            try out.print("# Minute {}\n", .{step_i + 1});

            if (opened.size == 0) {
                try out.print("- No valves are open.\n", .{});
            } else {
                switch (opened.size) {
                    1 => {
                        const k = opened.keyIterator().next() orelse @panic("should have 1");
                        try out.print("- Valve {} is open", .{k.*.name});
                    },
                    else => {
                        var keys = opened.keyIterator();
                        var i: usize = 0;
                        while (keys.next()) |k| {
                            nameList[i] = k.*.name;
                            i += 1;
                        }
                        var names = nameList[0..i];
                        std.sort.sort([]const u8, names, {}, memAsc(u8));

                        tmp.clearRetainingCapacity();

                        for (names) |name, j| {
                            if (j > 0) // the dread comma-and problem
                                try tmp.appendSlice(if (j < names.len - 1) ", " else if (j > 1) ", and " else " and ");
                            try tmp.appendSlice(name);
                        }

                        try out.print("- Valves {} are open", .{tmp.items});
                    },
                }
                try out.print(", releasing {} pressure.\n", .{totalOpen});
            }

            switch (step) {
                .move => |valve| {
                    try out.print("- You move to valve {}.\n", .{valve.name});
                },
                .open => |valve| {
                    try out.print("- You open valve {}.\n", .{valve.name});
                    opened.putAssumeCapacity(valve, {});
                    totalOpen += valve.flow;
                },
            }
            try out.print("\n", .{});
        }
    }

    try out.print(
        \\# Solution
        \\> {} total pressure released
        \\
    , .{
        result.totalReleased,
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
        return run(allocator, &bufin, output, config);

    var bufout = std.io.bufferedWriter(output.writer());
    try run(allocator, &bufin, &bufout, config);
    try bufout.flush();
    // TODO: sentinel-buffered output writer to flush lines progressively
    // ... may obviate the desire for raw / non-buffered output else
}
