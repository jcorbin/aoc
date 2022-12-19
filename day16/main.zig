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

        // Part 2 example
        .{
            .config = .{
                .verbose = 1,
                .max_steps = 26,
                .actor_names = &[_][]const u8{
                    "You",
                    "The elephant",
                },
            },
            .input = example_input,
            .expected = 
            \\# Minute 1
            \\- No valves are open.
            \\- You move to valve II.
            \\- The elephant moves to valve DD.
            \\
            \\# Minute 2
            \\- No valves are open.
            \\- You move to valve JJ.
            \\- The elephant opens valve DD.
            \\
            \\# Minute 3
            \\- Valve DD is open, releasing 20 pressure.
            \\- You open valve JJ.
            \\- The elephant moves to valve EE.
            \\
            \\# Minute 4
            \\- Valves DD and JJ are open, releasing 41 pressure.
            \\- You move to valve II.
            \\- The elephant moves to valve FF.
            \\
            \\# Minute 5
            \\- Valves DD and JJ are open, releasing 41 pressure.
            \\- You move to valve AA.
            \\- The elephant moves to valve GG.
            \\
            \\# Minute 6
            \\- Valves DD and JJ are open, releasing 41 pressure.
            \\- You move to valve BB.
            \\- The elephant moves to valve HH.
            \\
            \\# Minute 7
            \\- Valves DD and JJ are open, releasing 41 pressure.
            \\- You open valve BB.
            \\- The elephant opens valve HH.
            \\
            \\# Minute 8
            \\- Valves BB, DD, HH, and JJ are open, releasing 76 pressure.
            \\- You move to valve CC.
            \\- The elephant moves to valve GG.
            \\
            \\# Minute 9
            \\- Valves BB, DD, HH, and JJ are open, releasing 76 pressure.
            \\- You open valve CC.
            \\- The elephant moves to valve FF.
            \\
            \\# Minute 10
            \\- Valves BB, CC, DD, HH, and JJ are open, releasing 78 pressure.
            \\- The elephant moves to valve EE.
            \\
            \\# Minute 11
            \\- Valves BB, CC, DD, HH, and JJ are open, releasing 78 pressure.
            \\- The elephant opens valve EE.
            \\
            \\# Minute 12
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 13
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 14
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 15
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 16
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 17
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 18
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 19
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 20
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 21
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 22
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 23
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 24
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 25
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Minute 26
            \\- Valves BB, CC, DD, EE, HH, and JJ are open, releasing 81 pressure.
            \\
            \\# Solution
            \\> 1707 total pressure released
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

fn memLessThan(comptime T: type) fn (void, []const T, []const T) bool {
    const impl = struct {
        fn inner(_: void, a: []const T, b: []const T) bool {
            return std.mem.lessThan(T, a, b);
        }
    };
    return impl.inner;
}

pub const log_level: std.log.Level = .debug;

const log_others: std.log.Level = .debug;

pub fn log(
    comptime level: std.log.Level,
    comptime scope: @TypeOf(.EnumLiteral),
    comptime format: []const u8,
    args: anytype,
) void {
    const scope_prefix = "(" ++ switch (scope) {
        .default => @tagName(scope),
        else => if (@enumToInt(level) <= @enumToInt(log_others))
            @tagName(scope)
        else
            return,
    } ++ "): ";

    const prefix = "[" ++ comptime level.asText() ++ "] " ++ scope_prefix;

    // Print the message to stderr, silently ignoring any errors
    std.debug.getStderrMutex().lock();
    defer std.debug.getStderrMutex().unlock();
    const stderr = std.io.getStdErr().writer();
    nosuspend stderr.print(prefix ++ format ++ "\n", args) catch return;
}

const Parse = @import("parse.zig");
const Timing = @import("perf.zig").Timing(enum {
    parse,

    flowCostsRound,
    flowCosts,

    searchBatch,
    solve,

    report,
    overall,
});

const Config = struct {
    verbose: usize = 0,
    max_steps: usize = 30,
    start_name: []const u8 = "AA",
    actor_names: []const []const u8 = &[_][]const u8{
        "You",
    },
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

    pub fn parse(allocator: Allocator, reader: anytype) !World {
        var lines = Parse.lineScanner(reader);
        var builder = init: {
            var cur = try lines.next() orelse return error.NoInput;
            break :init Builder.initLine(allocator, cur) catch |err| return cur.carp(err);
        };
        while (try lines.next()) |cur|
            builder.parseLine(cur) catch |err| return cur.carp(err);
        // TODO per-line timings again
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
        const cap = @intCast(u32, slab.len);
        try valves.ensureTotalCapacity(self.arena.allocator(), cap);

        for (self.data.items) |item| switch (item) {
            .name => |name| if (name.len > 0) {
                var valve = &slab[0];
                slab = slab[1..];
                valve.* = .{
                    .name = name,
                    .next = links[0..0],
                };

                try valve.cost.ensureTotalCapacity(self.arena.allocator(), cap);

                try valves.put(self.arena.allocator(), name, valve);
            },
            else => {},
        };
        assert(slab.len == 0);

        {
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
        }

        // validate linkage
        {
            var any = false;
            for (valves.values()) |valve| {
                for (valve.next) |next| {
                    if (std.mem.indexOfScalar(*const Valve, next.next, valve) == null) {
                        std.log.warn("{s} -> {s} does not link back", .{ valve.name, next.name });
                        any = true;
                    }
                }
            }
            if (any) return error.InvalidGraphLinkage;
        }

        var openable = try self.arena.allocator().alloc(*const Valve, valves.entries.len);
        {
            var i: usize = 0;
            for (valves.values()) |valve| {
                if (valve.flow > 0) {
                    openable[i] = valve;
                    i += 1;
                }
            }
            openable = openable[0..i];
            std.sort.sort(*const Valve, openable, {}, Valve.flowLessThan);
        }

        return World{
            .allocator = self.allocator,
            .arena = self.arena,
            .valves = valves,
            .openable = openable,
        };
    }
};

const Valve = struct {
    const CostMap = std.AutoHashMapUnmanaged(*const Valve, usize);

    name: []const u8,
    flow: u16 = 0,
    next: []*Valve,
    cost: CostMap = .{},

    const Self = @This();

    pub fn flowLessThan(_: void, a: *const Valve, b: *const Valve) bool {
        return a.flow < b.flow;
    }

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
    const ValveMap = std.StringArrayHashMapUnmanaged(*Valve);

    allocator: Allocator,
    arena: std.heap.ArenaAllocator,
    valves: ValveMap,
    openable: []*const Valve,

    const Self = @This();

    pub fn deinit(self: *Self) void {
        self.arena.deinit();
    }
};

const search = @import("search.zig");

const Search = struct {
    const Self = @This();

    const Queue = search.Queue(
        *Plan,
        Expander,
        *Self,
        comparePlan,
        expandPlan,
        consumePlan,
        destroyPlan,
    );

    allocator: Allocator,
    world: *const World,
    result: ?*Plan = null,

    fn deinit(self: *Self) void {
        if (self.result) |plan| {
            self.result = null;
            self.destroyPlan(plan);
        }
    }

    fn consumePlan(self: *Self, plan: *Plan) search.Action {
        if (plan.availableSteps() > 0) {
            if (self.result) |prior| {
                if (plan.potentialReleased() <= prior.totalReleased)
                    return .skip;
            }
            return .queue;
        } else {
            if (self.result) |prior| {
                if (plan.totalReleased <= prior.totalReleased)
                    return .skip;
                self.destroyPlan(prior);
            }
            self.result = plan;
            return .take;
        }
    }

    fn destroyPlan(self: *Self, plan: *Plan) void {
        plan.deinit(self.allocator);
    }

    fn comparePlan(_: *Self, a: *Plan, b: *Plan) std.math.Order {
        // NOTE .lt -> a expandBefore b

        switch (std.math.order(a.availableSteps(), b.availableSteps())) {

            // explore better plans first
            .eq => return std.math.order(a.totalReleased, b.totalReleased).invert(),

            // deepen paths first to keep search frontier smaller
            else => |ord| return ord,
        }
    }

    fn expandPlan(self: *Self, plan: *Plan) Expander {
        var count: usize = 1;

        var open_remain = plan.openable.len;
        var idle_actors: usize = 0;
        for (plan.actors) |actor| {
            if (actor.state == .idle) idle_actors += 1;
        }

        if (open_remain > 0 and idle_actors > 0) {
            var actor_order = idle_actors;
            while (actor_order > 1) : (actor_order -= 1)
                count *= actor_order;

            while (open_remain > 0 and idle_actors > 0) {
                count *= open_remain;
                open_remain -= 1;
                idle_actors -= 1;
            }
        }

        return .{
            .self = self,
            .prior = plan,
            .count = count,
        };
    }

    const Expander = struct {
        self: *Self,
        prior: *Plan,
        reused: bool = false,

        count: usize,
        i: usize = 0,

        pub fn next(it: *@This()) !?*Plan {
            const i = it.i;
            if (it.reused or i >= it.count) return null;
            it.i += 1;

            var open_remain = it.prior.openable.len;
            if (open_remain == 0) {
                var plan = try it.reuse();
                while (plan.len < plan.max)
                    for (plan.actors) |*actor|
                        try plan.noop(actor);
                return plan;
            }

            var idle_actors: usize = 0;
            for (it.prior.actors) |actor| {
                if (actor.state == .idle) idle_actors += 1;
            }

            var skip = false;
            var j = i;

            // TODO reuse is sus given the need to read prior state in the indexing loop below
            // var plan = if (it.i == it.count)
            //     try it.reuse()
            // else
            //     try it.prior.clone(it.self.allocator);
            // errdefer if (plan != it.prior)
            //     plan.deinit(it.self.allocator);

            var plan = try it.prior.clone(it.self.allocator);
            errdefer plan.deinit(it.self.allocator);

            // shuffle idle actors
            var ia: usize = idle_actors;
            each_idle_actor: while (ia > 0 and open_remain > 0) : (ia -= 1) {
                const idle_actor_i = j % ia;
                j /= ia;

                const actor_id = choose_actor: {
                    var k = idle_actor_i;
                    for (it.prior.actors) |actor, id| if (actor.state == .idle) {
                        if (k == 0) break :choose_actor id;
                        k -= 1;
                    };
                    skip = true;
                    break :each_idle_actor;
                };

                // shuffle openable valves, but don't double-assign
                const open_i = j % open_remain;
                j /= open_remain;
                open_remain -= 1;
                const open = it.prior.openable[open_i];

                for (it.prior.actors) |other, other_id| {
                    if (actor_id != other_id and
                        other.state == .open and
                        other.state.open == open)
                        continue :each_idle_actor;
                }
                plan.actors[actor_id].state = .{ .open = open };
            }
            // now all actors are non-idle (if possible)

            // run actors until plan is full or at least one actor transitions back to idle
            while (!skip and plan.len < plan.max) {
                var any_idled = false;
                for (plan.actors) |*actor| switch (actor.state) {
                    .idle => try plan.noop(actor),
                    .open => |goal| {
                        if (actor.at != goal) {
                            if (actor.bestNextMove(goal)) |next_move| {
                                try plan.moveTo(actor, next_move);
                                continue;
                            }
                        } else if (plan.canOpen(actor.at)) {
                            try plan.open(actor);
                            actor.state = .{ .idle = {} };
                            any_idled = true;
                            continue;
                        }

                        try plan.noop(actor);
                        skip = true;
                        break;
                    },
                };
                if (any_idled) break;
            }

            if (skip) {
                const last = plan == it.prior;
                plan.deinit(it.self.allocator);
                return if (last) null else it.next();
            }

            return plan;
        }

        pub fn reuse(it: *@This()) !*Plan {
            if (it.reused) return error.AlreadyReused;
            it.reused = true;
            return it.prior;
        }

        pub fn deinit(it: *@This()) void {
            if (!it.reused)
                it.prior.deinit(it.self.allocator);
        }
    };
};

const ValveSet = std.AutoHashMapUnmanaged(*const Valve, void);

const Plan = struct {
    const Step = union(enum) {
        noop: void,
        move: *const Valve,
        open: *const Valve,
    };

    const Steps = std.ArrayListUnmanaged(Step);

    const Actor = struct {
        at: *const Valve,
        state: union(enum) {
            idle: void,
            open: *const Valve,
        } = .{ .idle = {} },

        steps: Steps,

        pub fn initCapacity(allocator: Allocator, at: *const Valve, cap: usize) !Actor {
            var steps = try Steps.initCapacity(allocator, cap);
            errdefer steps.deinit(allocator);

            return Actor{
                .at = at,
                .steps = steps,
            };
        }

        pub fn clone(self: Actor, allocator: Allocator) !Actor {
            var other = self;

            other.steps = try self.steps.clone(allocator);
            errdefer other.steps.deinit(allocator);

            return other;
        }

        pub fn deinit(self: *Actor, allocator: Allocator) void {
            self.steps.deinit(allocator);
            self.* = undefined;
        }

        pub fn availableSteps(actor: @This()) usize {
            return actor.steps.capacity - actor.steps.items.len;
        }

        pub fn bestNextMove(actor: @This(), goal: *const Valve) ?*const Valve {
            var best: ?*const Valve = null;
            var best_cost: usize = 0;
            for (actor.at.next) |next_valve| {
                const cost = next_valve.cost.get(goal) orelse continue;
                if (best == null or cost < best_cost) {
                    best = next_valve;
                    best_cost = cost;
                }
            }
            return best;
        }
    };

    op: []*const Valve,
    openable: []*const Valve,

    actors: []Actor,
    len: usize = 0,
    max: usize,

    totalOpen: usize = 0,
    nextOpen: usize = 0,
    totalReleased: usize = 0,

    const Self = @This();

    pub fn init(allocator: Allocator, params: struct {
        start: *Valve,
        max_steps: usize,
        openable: []*const Valve,
        actor_names: []const []const u8,
    }) !*Self {
        var self = try allocator.create(Self);
        errdefer allocator.destroy(self);

        var op = try allocator.dupe(*const Valve, params.openable);
        errdefer allocator.free(op);

        var actors = try allocator.alloc(Actor, params.actor_names.len);
        errdefer allocator.free(actors);

        var actor_i: usize = 0;
        errdefer while (actor_i > 0) {
            actor_i -= 1;
            actors[actor_i].deinit(allocator);
        };

        for (params.actor_names) |_| {
            actors[actor_i] = try Actor.initCapacity(allocator, params.start, params.max_steps);
            actor_i += 1;
        }

        self.* = Self{
            .op = op,
            .openable = op,
            .max = params.max_steps,
            .actors = actors,
        };

        return self;
    }

    pub fn clone(self: Self, allocator: Allocator) !*Self {
        var other = try allocator.create(Self);
        errdefer allocator.destroy(other);
        other.* = self;

        other.openable = try allocator.dupe(*const Valve, self.openable);
        errdefer allocator.free(other.openable);
        other.op = other.openable;

        var actors = try allocator.alloc(Actor, self.actors.len);
        errdefer allocator.free(actors);
        other.actors = actors;

        var actor_i: usize = 0;
        errdefer while (actor_i > 0) {
            actor_i -= 1;
            actors[actor_i].deinit(allocator);
        };

        for (self.actors) |actor| {
            actors[actor_i] = try actor.clone(allocator);
            actor_i += 1;
        }

        return other;
    }

    pub fn deinit(self: *Self, allocator: Allocator) void {
        allocator.free(self.op);
        for (self.actors) |*actor|
            actor.deinit(allocator);
        allocator.free(self.actors);
        self.* = undefined;
        allocator.destroy(self);
    }

    pub fn canOpen(self: Self, valve: *const Valve) bool {
        return std.mem.indexOfScalar(*const Valve, self.openable, valve) != null;
    }

    pub fn noop(self: *Self, actor: *Actor) !void {
        if (actor.availableSteps() == 0) return error.PlanFull;
        if (actor.steps.items.len > self.len) return error.AlreadyActed;
        actor.steps.appendAssumeCapacity(.{ .noop = {} });
        self.checkActors();
    }

    pub fn moveTo(self: *Self, actor: *Actor, move: *const Valve) !void {
        if (actor.availableSteps() == 0) return error.PlanFull;
        if (actor.steps.items.len > self.len) return error.AlreadyActed;
        actor.at = move;
        actor.steps.appendAssumeCapacity(.{ .move = move });
        self.checkActors();
    }

    pub fn open(self: *Self, actor: *Actor) !void {
        if (actor.availableSteps() == 0) return error.PlanFull;
        if (actor.steps.items.len > self.len) return error.AlreadyActed;
        const i = std.mem.indexOfScalar(*const Valve, self.openable, actor.at) orelse
            return error.CantOpen;
        const valve = actor.at;
        assert(valve.flow > 0);
        actor.steps.appendAssumeCapacity(.{ .open = valve });
        std.mem.copy(*const Valve, self.openable[i..], self.openable[i + 1 ..]);
        self.openable = self.openable[0 .. self.openable.len - 1];
        self.nextOpen += valve.flow;
        self.checkActors();
    }

    pub fn checkActors(self: *Self) void {
        for (self.actors) |*actor|
            if (actor.steps.items.len <= self.len) return;
        self.len += 1;
        self.totalReleased += self.totalOpen;
        self.totalOpen += self.nextOpen;
        self.nextOpen = 0;
    }

    pub fn availableSteps(self: Self) usize {
        return self.max - self.len;
    }

    pub fn potentialReleased(self: Self) usize {
        const stepsRem = self.availableSteps();
        const willAccum = self.totalOpen * stepsRem;

        var canAccum: usize = 0;
        var couldOpen: usize = 0;
        var step: usize = self.len;

        for (self.actors) |*actor| {
            if (self.canOpen(actor.at)) couldOpen += actor.at.flow;
        }

        var oi: usize = 1;
        var ai: usize = 0;
        var nextOpen: usize = 0;
        model_step: while (step < self.max) {
            if (oi > self.openable.len) break;

            const op = self.openable[self.openable.len - oi];
            oi += 1;

            for (self.actors) |*actor| {
                if (op == actor.at) continue :model_step;
            }

            nextOpen += op.flow;
            ai += 1;

            var flush = false;
            if (ai > self.actors.len) {
                ai = 0;
                flush = true;
            } else if (oi >= self.openable.len) {
                flush = true;
            }

            if (flush) {
                // move step
                canAccum += couldOpen;
                step += 1;
                if (step >= self.max) break;

                // open step
                canAccum += couldOpen;
                couldOpen += nextOpen;
                nextOpen = 0;
                step += 1;
            }
        }

        if (nextOpen > 0 and step < self.max) {
            // move step
            canAccum += couldOpen;
            step += 1;
            if (step < self.max) {
                canAccum += couldOpen;
                couldOpen += nextOpen;
                nextOpen = 0;
                step += 1;
            }
        }

        canAccum += couldOpen * (self.max - step);

        return self.totalReleased + willAccum + canAccum;
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

    {
        var roundTime = try timing.timer(.flowCostsRound);

        // init cost=1 for self and cost=1 for next
        for (world.valves.values()) |valve| {
            for (valve.next) |next| {
                var res = valve.cost.getOrPutAssumeCapacity(next);
                if (!res.found_existing)
                    res.value_ptr.* = 1;
            }
            valve.cost.putAssumeCapacity(valve, 0);
        }
        try roundTime.lap();

        // transitively close cost tables, keeping min cost at each valve
        while (true) {
            var any = false;
            for (world.valves.values()) |valve| {
                for (valve.next) |next| {
                    var it = next.cost.iterator();
                    while (it.next()) |entry| {
                        var res = valve.cost.getOrPutAssumeCapacity(entry.key_ptr.*);
                        var cost = 1 + entry.value_ptr.*;
                        if (!res.found_existing or res.value_ptr.* > cost) {
                            res.value_ptr.* = cost;
                            any = true;
                        }
                    }
                }
            }
            try roundTime.lap();
            if (!any) break;
        }
    }
    try timing.markPhase(.flowCosts);

    // validate costs
    {
        var any = false;
        for (world.valves.values()) |valve| {
            var it = valve.cost.iterator();
            while (it.next()) |entry| {
                const cost = entry.value_ptr.*;
                const next = entry.key_ptr.*;
                if (next == valve) {
                    if (cost != 0) {
                        std.log.warn("{s} self cost {} != 0", .{ valve.name, cost });
                        any = true;
                    }
                } else {
                    const cocost = next.cost.get(valve) orelse 0;
                    if (cost != cocost) {
                        std.log.warn("{s} <- {} != {} -> {s} asymmetric cost", .{ valve.name, cost, cocost, next.name });
                        any = true;
                    }
                }
            }
        }
        if (any) return error.InvalidGraphCosts;
    }

    var srch = Search{
        .allocator = allocator, // NOTE use to expose leaks under test
        // .allocator = arena.allocator(),
        .world = &world,
    };
    defer srch.deinit();

    var queue = Search.Queue.init(srch.allocator, &srch);
    defer queue.deinit();

    try queue.add(try Plan.init(srch.allocator, .{
        .openable = world.openable,
        .start = world.valves.get(config.start_name) orelse return error.NoStartValve,
        .max_steps = config.max_steps,
        .actor_names = config.actor_names,
    }));

    while (!queue.done()) {
        var roundTime = try timing.timer(.searchBatch);
        const ran = try queue.runUpto(1_000);
        try roundTime.lap();

        const time = timing.data.items[timing.data.items.len - 1].time;
        const best = if (srch.result) |res| res.totalReleased else 0;
        std.log.info("searched {} in {} ; best = {} depth = {}", .{
            ran,
            time,
            best,
            queue.queue.len,
        });
    }

    // NOTE the final result can have redundant final moves while some actors
    //      continue to move while the critical finisher(s) do also
    // TODO These moves could be pruned away by an optimizer at this point
    const result = srch.result orelse return error.NoResultFound;
    try timing.markPhase(.solve);

    if (config.verbose > 0) {
        var opened = ValveSet{};
        try opened.ensureUnusedCapacity(arena.allocator(), @intCast(u32, result.max));
        var totalOpen: usize = 0;
        var nameList = try arena.allocator().alloc([]const u8, result.max);
        var tmp = std.ArrayList(u8).init(arena.allocator());

        var step_i: usize = 0;
        while (step_i < result.max) : (step_i += 1) {
            try out.print("# Minute {}\n", .{step_i + 1});

            if (opened.size == 0) {
                try out.print("- No valves are open.\n", .{});
            } else {
                var keys = opened.keyIterator();
                switch (opened.size) {
                    1 => {
                        const k = keys.next() orelse @panic("should have 1");
                        try out.print("- Valve {s} is open", .{k.*.name});
                    },
                    else => {
                        var i: usize = 0;
                        while (keys.next()) |k| {
                            nameList[i] = k.*.name;
                            i += 1;
                        }
                        var names = nameList[0..i];
                        std.sort.sort([]const u8, names, {}, memLessThan(u8));

                        tmp.clearRetainingCapacity();
                        for (names) |name, j| {
                            if (j > 0) // the dread comma-and problem
                                try tmp.appendSlice(if (j < names.len - 1) ", " else if (j > 1) ", and " else " and ");
                            try tmp.appendSlice(name);
                        }

                        try out.print("- Valves {s} are open", .{tmp.items});
                    },
                }
                try out.print(", releasing {} pressure.\n", .{totalOpen});
            }

            for (result.actors) |actor, ai| {
                const actor_name = config.actor_names[ai];
                const itsMe = std.mem.eql(u8, actor_name, "You");
                const step = actor.steps.items[step_i];
                switch (step) {
                    .noop => {},
                    .move => |valve| {
                        if (itsMe)
                            try out.print("- You move to valve {s}.\n", .{valve.name})
                        else
                            try out.print("- {s} moves to valve {s}.\n", .{ actor_name, valve.name });
                    },
                    .open => |valve| {
                        if (itsMe)
                            try out.print("- You open valve {s}.\n", .{valve.name})
                        else
                            try out.print("- {s} opens valve {s}.\n", .{ actor_name, valve.name });
                        opened.putAssumeCapacity(valve, {});
                        totalOpen += valve.flow;
                    },
                }
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
    .enable_memory_limit = true,
    .verbose_log = false,
});

var gpa = MainAllocator{};

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
                    \\  -e or
                    \\  --elephant
                    \\    train an elephant for part 2
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
            } else if (arg.is(.{ "-e", "--elephant" })) {
                // TODO we could support N elephants at this point, but that
                //      wasn't necessary for Part 2
                config.actor_names = &[_][]const u8{
                    "You",
                    "The elephant",
                };
                config.max_steps = 26;
            } else if (arg.is(.{ "-v", "--verbose" })) {
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
    }) catch |err| switch (err) {
        mem.Allocator.Error.OutOfMemory => {
            const have_leaks = gpa.detectLeaks();
            std.log.err("{} detectLeaks() -> {}", .{ err, have_leaks });
            std.process.exit(9);
        },
        else => return err,
    };
    if (gpa.detectLeaks()) std.process.exit(9);
}
