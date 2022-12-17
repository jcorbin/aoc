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

// TODO const Plan = struct {};

// TODO build a search.zig module based on Day XXX

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

    // TODO: search for a solution
    // TODO: if verbose, print solution playback
    var start = world.valves.get("AA") orelse return error.MissingStartValve;
    try out.print("TODO start from {}\n", .{start});

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
