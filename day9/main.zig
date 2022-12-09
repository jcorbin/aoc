const std = @import("std");
const Allocator = std.mem.Allocator;

test "example" {
    const example =
        \\R 4
        \\U 4
        \\L 3
        \\D 1
        \\R 4
        \\D 1
        \\L 5
        \\R 2
        \\
    ;

    const expected =
        \\# Part 1
        \\
        \\```
        \\..##..
        \\...##.
        \\.####.
        \\....#.
        \\s###..
        \\```
        \\
        \\> 13
        \\
    ;

    const allocator = std.testing.allocator;

    var input = std.io.fixedBufferStream(example);
    var output = std.ArrayList(u8).init(allocator);
    defer output.deinit();

    run(allocator, &input, &output) catch |err| {
        std.debug.print("```pre-error output:\n{s}\n```\n", .{output.items});
        return err;
    };
    try std.testing.expectEqualStrings(expected, output.items);
}

const Parse = @import("./parse.zig");
const Timing = @import("./perf.zig").Timing;

fn run(
    allocator: Allocator,

    // TODO: better "any .reader()-able / any .writer()-able" interfacing
    input: anytype,
    output: anytype,
) !void {
    var timing = try Timing(enum {
        parse,
        parseLine,
        part1,
        part2,
        overall,
    }).start(allocator);
    defer timing.deinit();
    defer timing.printDebugReport();

    // FIXME: uncomment this if solutions need heap memory below
    // var arena = std.heap.ArenaAllocator.init(allocator);
    // defer arena.deinit();

    var lines = Parse.lineScanner(input.reader());
    var out = output.writer();

    // FIXME: such computer
    var lineTime = try std.time.Timer.start();
    while (try lines.next()) |*cur| {
        _ = cur; // FIXME: much line
        try timing.collect(.parseLine, lineTime.lap());
    }
    try timing.markPhase(.parse);

    // FIXME: measure any other distinct computation phases before part1/part2 particulars

    // FIXME: solve part 1
    {
        try out.print("# Part 1\n", .{});
        try out.print("> {}\n", .{42});
    }
    try timing.markPhase(.part1);

    // FIXME: solve part 2
    {
        try out.print("\n# Part 2\n", .{});
        try out.print("> {}\n", .{42});
    }
    try timing.markPhase(.part2);

    try timing.finish(.overall);
}

pub fn main() !void {
    const allocator = std.heap.page_allocator;

    var input = std.io.getStdIn();
    var output = std.io.getStdOut();

    var bufin = std.io.bufferedReader(input.reader());
    var bufout = std.io.bufferedWriter(output.writer());

    try run(allocator, &bufin, &bufout);
    try bufout.flush();

    // TODO: argument parsing to steer input selection

    // TODO: sentinel-buffered output writer to flush lines progressively

    // TODO: input, output, and run-time metrics
}
