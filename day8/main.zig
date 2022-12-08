const std = @import("std");
const Allocator = std.mem.Allocator;

test "example" {
    const example =
        \\such data
        \\30373
        \\25512
        \\65332
        \\33549
        \\35390
        \\
    ;

    // In this example, that only leaves the interior nine trees to consider:
    //
    // - The top-left 5 is visible from the left and top.
    //   (It isn't visible from the right or bottom since other trees of height 5 are in the way.)
    // - The top-middle 5 is visible from the top and right.
    // - The top-right 1 is not visible from any direction;
    //   for it to be visible, there would need to only be trees of height 0 between it and an edge.
    // - The left-middle 5 is visible, but only from the right.
    // - The center 3 is not visible from any direction;
    //   for it to be visible, there would need to be only trees of at most height 2 between it and an edge.
    // - The right-middle 3 is visible from the right.
    // - In the bottom row, the middle 5 is visible, but the 3 and 4 are not.
    //
    // With `16` trees visible on the edge and another 5 visible in the
    // interior, a total of 21 trees are visible in this arrangement.

    const expected =
        \\> 21
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
    var timing = Timing(enum {
        parse,
        parseLine,
        part1,
        part2,
        overall,
    }).init(allocator);
    defer timing.deinit();
    var runTime = try std.time.Timer.start();
    var phaseTime = runTime;

    var lines = Parse.lineScanner(input.reader());
    var out = output.writer();

    // FIXME: such computer
    var lineTime = try std.time.Timer.start();
    while (try lines.next()) |*cur| {
        _ = cur; // FIXME: much line
        try timing.collect(.parseLine, lineTime.lap());
    }
    try timing.collect(.parseAll, phaseTime.lap());

    // FIXME: measure any other distinct computation phases before part1/part2 particulars

    try out.print("# Part 1\n", .{});
    // FIXME solve
    try timing.collect(.part1, phaseTime.lap());
    try out.print("> {}\n", .{42});

    try out.print("\n# Part 2\n", .{});
    // FIXME solve, then:
    try timing.collect(.part2, phaseTime.lap());
    try out.print("> {}\n", .{42});

    try timing.collect(.overall, runTime.lap());

    std.debug.print("# Timing\n\n", .{});
    for (timing.data.items) |item| {
        if (item.tag != .parseLine) {
            std.debug.print("- {} {}\n", .{ item.time, item.tag });
        }
    }
    std.debug.print("\n", .{});
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
