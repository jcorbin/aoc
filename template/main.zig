const std = @import("std");
const Allocator = std.mem.Allocator;

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
            \\such data
            \\
            ,
            .expected = 
            \\# Solution
            \\> 42
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

const Parse = @import("./parse.zig");
const Timing = @import("./perf.zig").Timing;

const Config = struct {
    verbose: bool = false,
};

fn parseLine(cur: *Parse.Cursor) !void {
    // TODO refactor as needed, e.g. onto a stateful builder struct
    try cur.expectEnd(error.ParseLineNotImplemented);
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
        solve,
        overall,
    }).start(allocator);
    defer timing.deinit();
    defer timing.printDebugReport();

    var out = output.writer();

    // FIXME: hookup your config
    _ = config;

    // FIXME: uncomment this if solutions need heap memory below
    // var arena = std.heap.ArenaAllocator.init(allocator);
    // defer arena.deinit();

    { // FIXME: parse input (store intermediate form, or evaluate)
        var lines = Parse.lineScanner(input.reader());
        while (try lines.next()) |*cur| {
            var lineTime = try std.time.Timer.start();
            parseLine(cur) catch |err| {
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

        try timing.markPhase(.parse);
    }

    // FIXME: solve...
    {
        try out.print(
            \\# Solution
            \\> {}
            \\
        , .{
            42,
        });
    }
    try timing.markPhase(.solve);

    try timing.finish(.overall);
}

const ArgParser = @import("./args.zig").Parser;

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
