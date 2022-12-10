const std = @import("std");
const Allocator = std.mem.Allocator;

test "example" {
    const test_cases = [_]struct {
        input: []const u8,
        expected: []const u8,
        config: Config,
    }{
        // Part 1 small example
        .{
            .config = .{
                .verbose = true,
            },
            .input = 
            \\noop
            \\addx 3
            \\addx -5
            \\
            ,
            .expected = 
            \\# Eval 1. `noop`
            \\    cycle: 1 x: 1
            \\# Eval 2. `addx 3`
            \\    cycle: 2 x: 1
            \\    cycle: 3 x: 1
            \\# Eval 3. `addx -5`
            \\    cycle: 4 x: 4
            \\    cycle: 5 x: 4
            \\# Halt
            \\    cycle: 6 x: -1
            \\
            ,
        },
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

const CPU = struct {
    cycle: usize = 1,
    x: i64 = 1,

    const Op = union(enum) {
        noop: void,
        addx: i64,

        pub fn parse(buf: []const u8) !Op {
            return if (std.mem.eql(u8, buf, "noop"))
                Op{ .noop = {} }
            else if (std.mem.startsWith(u8, buf, "addx "))
                Op{ .addx = try std.fmt.parseInt(i64, buf[5..], 10) }
            else
                error.UnrecognizedOp;
        }
    };

    const Iterator = struct {
        from: CPU,
        to: CPU,

        pub fn next(it: *@This()) ?CPU {
            if (it.from.cycle < it.to.cycle) {
                const prior = it.from;
                it.from.cycle += 1;
                return prior;
            }
            return null;
        }
    };

    const Self = @This();

    pub fn tick(self: Self, op: Op) Iterator {
        return .{ .from = self, .to = self.exec(op) };
    }

    pub fn exec(self: Self, op: Op) Self {
        switch (op) {
            .noop => return .{ .cycle = self.cycle + 1 },
            .addx => |n| return .{ .cycle = self.cycle + 2, .x = self.x + n },
        }
    }
};

const Parse = @import("./parse.zig");
const Timing = @import("./perf.zig").Timing;

const Config = struct {
    verbose: bool = false,
};

fn run(
    allocator: Allocator,

    // TODO: better "any .reader()-able / any .writer()-able" interfacing
    input: anytype,
    output: anytype,
    config: Config,
) !void {
    var timing = try Timing(enum {
        eval,
        evalLine,
        evalLineVerbose,
        report,
        overall,
    }).start(allocator);
    defer timing.deinit();
    defer timing.printDebugReport();

    // FIXME: uncomment this if solutions need heap memory below
    // var arena = std.heap.ArenaAllocator.init(allocator);
    // defer arena.deinit();

    var lines = Parse.lineScanner(input.reader());
    var out = output.writer();

    var cpu = CPU{};

    // evaluate input
    while (try lines.next()) |*cur| {
        var lineTime = try std.time.Timer.start();
        const op = try CPU.Op.parse(cur.buf);

        if (config.verbose) try out.print(
            \\# Eval {}. `{s}`
            \\
        , .{ cur.count, cur.buf });

        var ticks = cpu.tick(op);
        while (ticks.next()) |state| {
            if (config.verbose) try out.print(
                \\    cycle: {} x: {}
                \\
            , .{ state.cycle, state.x });
            // TODO collect signal strength
        }
        cpu = ticks.to;

        try timing.collect(.evalLine, lineTime.lap());
    }
    try timing.markPhase(.eval);

    // report
    // TODO signal strength sum
    try out.print(
        \\# Halt
        \\    cycle: {} x: {}
        \\
    , .{ cpu.cycle, cpu.x });
    try timing.markPhase(.report);

    try timing.finish(.overall);
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
