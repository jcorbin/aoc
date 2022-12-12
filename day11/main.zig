const std = @import("std");
const Allocator = std.mem.Allocator;

test "example" {
    const example_input =
        \\Monkey 0:
        \\  Starting items: 79, 98
        \\  Operation: new = old * 19
        \\  Test: divisible by 23
        \\    If true: throw to monkey 2
        \\    If false: throw to monkey 3
        \\
        \\Monkey 1:
        \\  Starting items: 54, 65, 75, 74
        \\  Operation: new = old + 6
        \\  Test: divisible by 19
        \\    If true: throw to monkey 2
        \\    If false: throw to monkey 0
        \\
        \\Monkey 2:
        \\  Starting items: 79, 60, 97
        \\  Operation: new = old * old
        \\  Test: divisible by 13
        \\    If true: throw to monkey 1
        \\    If false: throw to monkey 3
        \\
        \\Monkey 3:
        \\  Starting items: 74
        \\  Operation: new = old + 3
        \\  Test: divisible by 17
        \\    If true: throw to monkey 0
        \\    If false: throw to monkey 1
        \\
    ;

    const test_cases = [_]struct {
        input: []const u8,
        expected: []const u8,
        config: Config,
    }{
        // Part 1 example: round 1 trace
        .{
            .config = .{
                .verbose = true,
                .trace = true,
                .rounds = 1,
            },
            .input = example_input,
            .expected = 
            \\# Trace
            \\    Monkey 0:
            \\      Monkey inspects an item with a worry level of 79.
            \\        Worry level is multiplied by 19 to 1501.
            \\        Monkey gets bored with item. Worry level is divided by 3 to 500.
            \\        Current worry level is not divisible by 23.
            \\        Item with worry level 500 is thrown to monkey 3.
            \\      Monkey inspects an item with a worry level of 98.
            \\        Worry level is multiplied by 19 to 1862.
            \\        Monkey gets bored with item. Worry level is divided by 3 to 620.
            \\        Current worry level is not divisible by 23.
            \\        Item with worry level 620 is thrown to monkey 3.
            \\    Monkey 1:
            \\      Monkey inspects an item with a worry level of 54.
            \\        Worry level increases by 6 to 60.
            \\        Monkey gets bored with item. Worry level is divided by 3 to 20.
            \\        Current worry level is not divisible by 19.
            \\        Item with worry level 20 is thrown to monkey 0.
            \\      Monkey inspects an item with a worry level of 65.
            \\        Worry level increases by 6 to 71.
            \\        Monkey gets bored with item. Worry level is divided by 3 to 23.
            \\        Current worry level is not divisible by 19.
            \\        Item with worry level 23 is thrown to monkey 0.
            \\      Monkey inspects an item with a worry level of 75.
            \\        Worry level increases by 6 to 81.
            \\        Monkey gets bored with item. Worry level is divided by 3 to 27.
            \\        Current worry level is not divisible by 19.
            \\        Item with worry level 27 is thrown to monkey 0.
            \\      Monkey inspects an item with a worry level of 74.
            \\        Worry level increases by 6 to 80.
            \\        Monkey gets bored with item. Worry level is divided by 3 to 26.
            \\        Current worry level is not divisible by 19.
            \\        Item with worry level 26 is thrown to monkey 0.
            \\    Monkey 2:
            \\      Monkey inspects an item with a worry level of 79.
            \\        Worry level is multiplied by itself to 6241.
            \\        Monkey gets bored with item. Worry level is divided by 3 to 2080.
            \\        Current worry level is divisible by 13.
            \\        Item with worry level 2080 is thrown to monkey 1.
            \\      Monkey inspects an item with a worry level of 60.
            \\        Worry level is multiplied by itself to 3600.
            \\        Monkey gets bored with item. Worry level is divided by 3 to 1200.
            \\        Current worry level is not divisible by 13.
            \\        Item with worry level 1200 is thrown to monkey 3.
            \\      Monkey inspects an item with a worry level of 97.
            \\        Worry level is multiplied by itself to 9409.
            \\        Monkey gets bored with item. Worry level is divided by 3 to 3136.
            \\        Current worry level is not divisible by 13.
            \\        Item with worry level 3136 is thrown to monkey 3.
            \\    Monkey 3:
            \\      Monkey inspects an item with a worry level of 74.
            \\        Worry level increases by 3 to 77.
            \\        Monkey gets bored with item. Worry level is divided by 3 to 25.
            \\        Current worry level is not divisible by 17.
            \\        Item with worry level 25 is thrown to monkey 1.
            \\      Monkey inspects an item with a worry level of 500.
            \\        Worry level increases by 3 to 503.
            \\        Monkey gets bored with item. Worry level is divided by 3 to 167.
            \\        Current worry level is not divisible by 17.
            \\        Item with worry level 167 is thrown to monkey 1.
            \\      Monkey inspects an item with a worry level of 620.
            \\        Worry level increases by 3 to 623.
            \\        Monkey gets bored with item. Worry level is divided by 3 to 207.
            \\        Current worry level is not divisible by 17.
            \\        Item with worry level 207 is thrown to monkey 1.
            \\      Monkey inspects an item with a worry level of 1200.
            \\        Worry level increases by 3 to 1203.
            \\        Monkey gets bored with item. Worry level is divided by 3 to 401.
            \\        Current worry level is not divisible by 17.
            \\        Item with worry level 401 is thrown to monkey 1.
            \\      Monkey inspects an item with a worry level of 3136.
            \\        Worry level increases by 3 to 3139.
            \\        Monkey gets bored with item. Worry level is divided by 3 to 1046.
            \\        Current worry level is not divisible by 17.
            \\        Item with worry level 1046 is thrown to monkey 1.
            \\
            \\# Round 1
            \\    Monkey 0: 20, 23, 27, 26
            \\    Monkey 1: 2080, 25, 167, 207, 401, 1046
            \\    Monkey 2:
            \\    Monkey 3:
            \\
            ,
        },
        // TODO Part 1 example: round state for 1-10,15,20

        // TODO Part 1 example: outcome
        // In this example, the two most active monkeys inspected items 101 and 105 times.
        // The level of *monkey business* in this situation can be found by multiplying
        // these together: *`10605`*.
    };

    const allocator = std.testing.allocator;

    for (test_cases) |tc| {
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

const Item = struct {
    worry: u32,
};

/// A simple math expression on a single variable X
const Op = union(enum) {
    x: void,
    value: u32,
    add: [2]*@This(),
    mul: [2]*@This(),

    const Self = @This();

    pub fn eval(self: Self, x: u32) u32 {
        return switch (self) {
            .x => x,
            .value => |n| n,
            .add => |expr| expr[0].eval(x) + expr[1].eval(x),
            .mul => |expr| expr[0].eval(x) * expr[1].eval(x),
        };
    }

    const TraceWriter = std.ArrayList(u8).Writer;

    pub fn trace(self: Self, x: u32, tw: TraceWriter) u32 {
        switch (self) {
            .x => {
                tw.print("Worry level", .{}) catch {};
                return x;
            },
            .value => |n| {
                tw.print("{d}", .{n}) catch {};
                return n;
            },
            .add => |expr| {
                const a = expr[0].trace(x, tw);
                tw.print(" increases by ", .{}) catch {};
                const b = expr[1].retrace(x, tw);
                return a + b;
            },
            .mul => |expr| {
                const a = expr[0].trace(x, tw);
                tw.print(" is multiplied by ", .{}) catch {};
                const b = expr[1].retrace(x, tw);
                return a * b;
            },
        }
    }

    pub fn retrace(self: Self, x: u32, tw: TraceWriter) u32 {
        switch (self) {
            .x => {
                tw.print("itself", .{}) catch {};
                return x;
            },
            .value => |n| {
                tw.print("{d}", .{n}) catch {};
                return n;
            },
            .add => |expr| {
                const a = expr[0].retrace(x, tw);
                tw.print(" and increases by ", .{}) catch {};
                const b = expr[1].retrace(x, tw);
                return a + b;
            },
            .mul => |expr| {
                const a = expr[0].retrace(x, tw);
                tw.print(" and is multiplied by ", .{}) catch {};
                const b = expr[1].retrace(x, tw);
                return a * b;
            },
        }
    }
};

const Items = std.TailQueue(Item);

const ProtoMonkey = struct {
    id: usize,
    worldID: usize = 0,
    items: Items = .{},
    op: Op = .{ .x = {} },
    testDiv: u32 = 0,
    throwToID: [2]?usize = .{ null, null },
};

const Monkey = struct {
    id: usize,
    items: Items,
    op: Op,
    testDiv: u32,
    throwTo: [2]*Monkey,
};

const Parse = @import("./parse.zig");
const Timing = @import("./perf.zig").Timing;

const Config = struct {
    verbose: bool = false,
    trace: bool = false,
    rounds: usize = 0,
};

const MonkeyBuilder = struct {
    arena: *std.heap.ArenaAllocator,
    monkeys: std.AutoArrayHashMapUnmanaged(usize, *ProtoMonkey) = .{},
    monkey: ?*ProtoMonkey = null,
    inTest: bool = false,

    const Self = @This();

    pub fn compile(self: *Self) ![]Monkey {
        try self.flush();

        var allocator = self.arena.allocator();
        var monkeys = try allocator.alloc(Monkey, self.monkeys.entries.len);
        errdefer allocator.free(monkeys);

        // assign ID -> slot
        var worldID: usize = 0;
        for (self.monkeys.values()) |proto| {
            proto.worldID = worldID;
            worldID += 1;
        }

        // copy data and link
        for (self.monkeys.values()) |proto| {
            const throwIfID = proto.throwToID[0] orelse return error.NullThrowIf;
            const throwElseID = proto.throwToID[1] orelse return error.NullThrowElse;
            const protoIf = self.monkeys.get(throwIfID) orelse return error.MissingThrowIfID;
            const protoElse = self.monkeys.get(throwElseID) orelse return error.MissingThrowElseID;
            monkeys[proto.worldID] = .{
                .id = proto.id,
                .items = proto.items,
                .op = proto.op,
                .testDiv = proto.testDiv,
                .throwTo = .{
                    &monkeys[protoIf.worldID],
                    &monkeys[protoElse.worldID],
                },
            };
        }

        return monkeys;
    }

    pub fn flush(self: *Self) !void {
        var monkey = self.monkey orelse return;

        if (monkey.testDiv == 0) return error.MonkeyZeroTestDivisor;
        if (monkey.throwToID[0] == null) return error.MonkeyNullThrowIf;
        if (monkey.throwToID[1] == null) return error.MonkeyNullThrowElse;

        var allocator = self.arena.allocator();
        var res = try self.monkeys.getOrPut(allocator, monkey.id);
        if (res.found_existing) return error.MonkeyRedefined;
        res.value_ptr.* = monkey;

        self.monkey = null;
    }

    pub fn parseLine(self: *Self, cur: *Parse.Cursor) !void {
        var allocator = self.arena.allocator();

        if (std.mem.trim(u8, cur.buf, " ").len == 0)
            return self.flush();

        if (self.monkey) |monkey| {
            try cur.expectLiteral("  ", error.MonkeyExpectedIndent);

            if (cur.haveLiteral("Starting items:")) {
                while (cur.live()) {
                    cur.expectStar(' ');
                    var item = try allocator.create(Items.Node);
                    item.* = .{
                        .data = .{
                            .worry = try cur.expectInt(u32, 10, error.MonkeyItemExpected),
                        },
                    };
                    monkey.items.append(item);
                    if (cur.haveLiteral(",")) continue;
                    break;
                }
                try cur.expectEnd(error.MonkeyUnexpectedTrailer);
                return;
            }

            if (cur.haveLiteral("Operation:")) {
                cur.expectStar(' ');
                try cur.expectLiteral("new =", error.MonkeyOpExpectedNewEq);
                monkey.op = try self.parseOp(cur);
                try cur.expectEnd(error.MonkeyUnexpectedTrailer);
                return;
            }

            if (cur.haveLiteral("Test:")) {
                cur.expectStar(' ');
                try cur.expectLiteral("divisible by", error.MonkeyInvalidTest);
                cur.expectStar(' ');
                monkey.testDiv = try cur.expectInt(u32, 10, error.MonkeyExpectedTestDivisor);
                try cur.expectEnd(error.MonkeyUnexpectedTrailer);

                self.inTest = true;
                return;
            }

            if (self.inTest) {
                if (cur.haveLiteral("  ")) {
                    if (cur.haveLiteral("If")) {
                        cur.expectStar(' ');

                        if (cur.haveLiteral("true:")) {
                            cur.expectStar(' ');
                            try cur.expectLiteral("throw to monkey", error.MonkeyExpectedThrowTo);
                            cur.expectStar(' ');
                            monkey.throwToID[0] = try cur.expectInt(usize, 10, error.MonkeyExpectedID);
                        } else if (cur.haveLiteral("false:")) {
                            cur.expectStar(' ');
                            try cur.expectLiteral("throw to monkey", error.MonkeyExpectedThrowTo);
                            cur.expectStar(' ');
                            monkey.throwToID[1] = try cur.expectInt(usize, 10, error.MonkeyExpectedID);
                        } else return error.UnrecognizedMonkeyIfClause;

                        try cur.expectEnd(error.MonkeyUnexpectedTrailer);
                        return;
                    }
                } else self.inTest = false;
            }

            return error.UnrecognizedMonkeyLine;
        }

        try cur.expectLiteral("Monkey ", error.ExpectedMonkeyHeader);
        const id = try cur.expectInt(usize, 10, error.MonkeyHeaderExpectedID);
        try cur.expectLiteral(":", error.ExpectedMonkeyHeader);
        try cur.expectEnd(error.MonkeyUnexpectedTrailer);

        var monkey = try allocator.create(ProtoMonkey);
        monkey.* = .{ .id = id };
        self.monkey = monkey;
        self.inTest = false;
    }

    const ParseOpError = Allocator.Error || error{
        UnexpectedOpToken,
    };

    pub fn parseOp(self: *Self, cur: *Parse.Cursor) ParseOpError!Op {
        return self.parseOpTerm(cur, 0);
    }

    pub fn parseOpTerm(self: *Self, cur: *Parse.Cursor, prio: usize) ParseOpError!Op {
        cur.expectStar(' ');
        return if (cur.haveLiteral("old"))
            self.parseOpExpr(cur, .{ .x = {} }, prio)
        else if (cur.consumeInt(u32, 10)) |n|
            self.parseOpExpr(cur, .{ .value = n }, prio)
        else
            error.UnexpectedOpToken;
    }

    pub fn parseOpExpr(self: *Self, cur: *Parse.Cursor, left: Op, prio: usize) ParseOpError!Op {
        var allocator = self.arena.allocator();
        cur.expectStar(' ');
        switch (cur.peek() orelse return left) {
            '+' => {
                if (prio > 0) return left;
                cur.i += 1;
                var right = try self.parseOpTerm(cur, 0);
                var legs = try allocator.alloc(Op, 2);
                legs[0] = left;
                legs[1] = right;
                return Op{ .add = .{ &legs[0], &legs[1] } };
            },
            '*' => {
                if (prio > 1) return left;
                cur.i += 1;
                var right = try self.parseOpTerm(cur, 1);
                var legs = try allocator.alloc(Op, 2);
                legs[0] = left;
                legs[1] = right;
                return Op{ .mul = .{ &legs[0], &legs[1] } };
            },
            else => return error.UnexpectedOpToken,
        }
    }
};

fn World(
    comptime LogWriter: type,
) type {
    return struct {
        logWriter: LogWriter,

        traceEnabled: bool = false,
        traceOpen: bool = false,

        arena: std.heap.ArenaAllocator,
        monkeys: []Monkey = &[_]Monkey{},

        tmp: std.ArrayList(u8),

        const Self = @This();

        pub fn init(
            logWriter: LogWriter,
            allocator: Allocator,
        ) Self {
            return .{
                .logWriter = logWriter,
                .arena = std.heap.ArenaAllocator.init(allocator),
                .tmp = std.ArrayList(u8).init(allocator),
            };
        }

        pub fn deinit(self: *Self) void {
            self.tmp.deinit();
            self.arena.deinit();
        }

        pub fn trace(self: *Self, comptime fmt: []const u8, args: anytype) void {
            if (self.traceEnabled) {
                if (!self.traceOpen) {
                    self.logWriter.print("# Trace\n", .{}) catch return;
                    self.traceOpen = true;
                }
                self.logWriter.print("    " ++ fmt, args) catch return;
            }
        }

        pub fn run(self: *Self) void {
            for (self.monkeys) |*monkey| {
                if (self.traceEnabled) self.trace("Monkey {}:\n", .{monkey.id});

                while (monkey.items.popFirst()) |node| {
                    var worry = node.data.worry;

                    if (self.traceEnabled) {
                        self.trace("  Monkey inspects an item with a worry level of {}.\n", .{worry});

                        self.tmp.clearRetainingCapacity();
                        // try self.tmp.ensureTotalCapacity(1024);
                        var tmpW = self.tmp.writer();

                        worry = monkey.op.trace(worry, tmpW);
                        // worry = monkey.op.eval(worry);
                        tmpW.print(" to {}.", .{worry}) catch {};

                        self.trace("    {s}\n", .{self.tmp.items});
                    } else {
                        worry = monkey.op.eval(worry);
                    }

                    worry = @divTrunc(worry, 3);
                    if (self.traceEnabled)
                        self.trace("    Monkey gets bored with item. Worry level is divided by 3 to {}.\n", .{worry});

                    node.data.worry = worry;

                    const testValue = (worry % monkey.testDiv) == 0;

                    if (self.traceEnabled) {
                        if (testValue) {
                            self.trace("    Current worry level is divisible by {}.\n", .{monkey.testDiv});
                        } else {
                            self.trace("    Current worry level is not divisible by {}.\n", .{monkey.testDiv});
                        }
                    }

                    const toMonkey = monkey.throwTo[if (testValue) 0 else 1];

                    toMonkey.items.append(node);

                    if (self.traceEnabled)
                        self.trace("    Item with worry level {} is thrown to monkey {}.\n", .{ worry, toMonkey.id });
                }
            }
        }
    };
}

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
        run,
        runRound,
        solve,
        overall,
    }).start(allocator);
    defer timing.deinit();
    defer timing.printDebugReport();

    var out = output.writer();
    var world = World(@TypeOf(out)).init(out, allocator);
    defer world.deinit();

    world.traceEnabled = config.trace;

    var builder = MonkeyBuilder{ .arena = &world.arena };

    { // parse monkey definitions
        var lines = Parse.lineScanner(input.reader());
        while (try lines.next()) |*cur| {
            var lineTime = try std.time.Timer.start();
            builder.parseLine(cur) catch |err| {
                const space = " " ** 4096;
                std.debug.print(
                    \\Unable to parse line #{}:
                    \\> {s}
                    \\  {s}^-- {} here
                    \\
                , .{
                    cur.count,
                    cur.buf,
                    space[0..cur.i],
                    err,
                });
                return err;
            };
            try timing.collect(.parseLine, lineTime.lap());
        }

        world.monkeys = try builder.compile();

        try timing.markPhase(.parse);
    }

    { // run rounds
        var round: usize = 0;
        while (round < config.rounds) {
            var roundTime = try std.time.Timer.start();
            round += 1;
            world.run();

            world.traceOpen = false;
            try out.print(
                \\
                \\# Round {}
                \\
            , .{round});
            for (world.monkeys) |*monkey, i| {
                try out.print("    Monkey {}:", .{i});
                var item = monkey.items.first;
                while (item) |it| : (item = it.next) {
                    if (item == monkey.items.first) {
                        try out.print(" {}", .{it.data.worry});
                    } else {
                        try out.print(", {}", .{it.data.worry});
                    }
                }
                try out.print("\n", .{});
            }
            try timing.collect(.runRound, roundTime.lap());
        }
        try timing.markPhase(.run);
    }

    // // TODO: count monkey business
    // {
    //     try out.print(
    //         \\# Solution
    //         \\> {}
    //         \\
    //     , .{
    //         42,
    //     });
    // }
    // try timing.markPhase(.solve);

    try timing.finish(.overall);
}

fn collect(
    comptime T: type,
    allocator: std.mem.Allocator,
    qlist: std.TailQueue(T),
) !?[]T {
    var values = try allocator.alloc(T, qlist.len);
    var i: usize = 0;
    var it = qlist.first;
    while (it) |item| : (it = item.next) {
        values[i] = item.data;
        i += 1;
    }
    return values;
}

const ArgParser = @import("./args.zig").Parser;

pub fn main() !void {
    const allocator = std.heap.page_allocator;

    var input = std.io.getStdIn();
    var output = std.io.getStdOut();
    var config = Config{};

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
                    \\  -v or
                    \\  --verbose
                    \\    print world state after evaluating each input line
                    \\
                , .{args.progName()});
                std.process.exit(0);
            } else if (arg.is(.{ "-v", "--verbose" })) {
                config.verbose = true;
            } else return error.InvalidArgument;
        }
    }

    var bufin = std.io.bufferedReader(input.reader());
    var bufout = std.io.bufferedWriter(output.writer());

    try run(allocator, &bufin, &bufout, config);
    try bufout.flush();

    // TODO: sentinel-buffered output writer to flush lines progressively
}
