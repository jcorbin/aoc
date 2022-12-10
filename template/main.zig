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
            \\# Parse 1. `such data`
            \\# Solution
            \\> 42
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
        parse,
        parseLine,
        parseLineVerbose,
        solve,
        overall,
    }).start(allocator);
    defer timing.deinit();
    defer timing.printDebugReport();

    // FIXME: uncomment this if solutions need heap memory below
    // var arena = std.heap.ArenaAllocator.init(allocator);
    // defer arena.deinit();

    var lines = Parse.lineScanner(input.reader());
    var out = output.writer();

    // FIXME: parse input (store intermediate form, or evaluate)
    while (try lines.next()) |*cur| {
        var lineTime = try std.time.Timer.start();
        _ = cur; // FIXME: much line
        try timing.collect(.parseLine, lineTime.lap());

        if (config.verbose) {
            try out.print(
                \\# Parse {}. `{s}`
                \\
            , .{
                cur.count,
                cur.buf,
            });
            try timing.collect(.parseLineVerbose, lineTime.lap());
        }
    }
    try timing.markPhase(.parse);

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

pub fn main() !void {
    const allocator = std.heap.page_allocator;

    var input = std.io.getStdIn();
    var output = std.io.getStdOut();
    var config = Config{};

    // TODO more generic arg parsing; including input selection
    {
        var arena = std.heap.ArenaAllocator.init(allocator);
        defer arena.deinit();

        var args = try std.process.argsWithAllocator(arena.allocator());

        // TODO: make note of the program name for forming help strings?
        if (!args.skip()) return error.MissingArg0;

        while (args.next(arena.allocator())) |arg| {
            const flag = try arg;

            if (std.mem.eql(u8, flag, "-v") or
                std.mem.eql(u8, flag, "--verbose"))
            {
                config.verbose = true;
            } else return error.InvalidArgument;

            // TODO code fragments towards else:
            // if (std.mem.startsWith(u8, flag, "-"))
            // if (std.mem.eql(u8, flag, "--foo")) {
            //     const nextArg = args.next(arena.allocator()) orelse return error.MissingFlagValue;
            //     const value = try nextArg;
            //     config.foo = try std.fmt.parseInt(u4, value, 10);
            // }
        }
    }

    var bufin = std.io.bufferedReader(input.reader());
    var bufout = std.io.bufferedWriter(output.writer());

    try run(allocator, &bufin, &bufout, config);
    try bufout.flush();

    // TODO: sentinel-buffered output writer to flush lines progressively
}
