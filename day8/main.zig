const std = @import("std");
const Allocator = std.mem.Allocator;

test "example" {
    const example =
        \\30373
        \\25512
        \\65332
        \\33549
        \\35390
        \\
    ;

    // 3 0 3 7 3
    // 2 5 5 1 2
    // 6 5 3 3 2
    // 3 3 5 4 9
    // 3 5 3 9 0

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
        \\```visibility
        \\+ + + + +
        \\+ + + - +
        \\+ + - + +
        \\+ - + - +
        \\+ + + + +
        \\```
        \\
        \\# Part 1
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
    var timing = try Timing(enum {
        parse,
        parseLine,
        computeVisibility,
        displayVisibility,
        part1,
        part2,
        overall,
    }).start(allocator);
    defer timing.deinit();
    defer timing.printDebugReport();

    var arena = std.heap.ArenaAllocator.init(allocator);
    defer arena.deinit();

    var lines = Parse.lineScanner(input.reader());
    var out = output.writer();

    // parse height field input
    var lineTime = try std.time.Timer.start();
    var size: usize = 0;
    var height: []u4 = undefined;
    while (try lines.next()) |*cur| {
        if (size == 0) {
            size = cur.buf.len;
            height = try arena.allocator().alloc(u4, size * size);
        } else if (cur.buf.len != size) {
            return error.InvalidRowLength;
        }

        const y = cur.count - 1;
        for (cur.buf) |c, x| {
            if (c < '0' or c > '9') return error.InvalidDigit;
            const d = @intCast(u4, c - '0');
            height[size * y + x] = d;
        }

        try timing.collect(.parseLine, lineTime.lap());
    }
    if (size == 0) return error.NoInput;
    try timing.markPhase(.parse);

    // compute visibility
    var visible = try arena.allocator().alloc(bool, height.len);
    std.mem.set(bool, visible, false);

    {
        var n: usize = 1;
        while (n < size - 1) : (n += 1) {

            // march right
            {
                var x: usize = 0;
                var y: usize = n;
                visible[y * size + x] = true;
                var max = height[y * size + x];
                x += 1;
                while (x < size - 1) : (x += 1) {
                    const i = y * size + x;
                    const h = height[i];
                    if (h > max) {
                        max = h;
                        visible[i] = true;
                    }
                }
            }

            // march right
            {
                var x: usize = size - 1;
                var y: usize = n;
                visible[y * size + x] = true;
                var max = height[y * size + x];
                x -= 1;
                while (x > 0) : (x -= 1) {
                    const i = y * size + x;
                    const h = height[i];
                    if (h > max) {
                        max = h;
                        visible[i] = true;
                    }
                }
            }

            // march down
            {
                var x: usize = n;
                var y: usize = 0;
                visible[y * size + x] = true;
                var max = height[y * size + x];
                y += 1;
                while (y < size - 1) : (y += 1) {
                    const i = y * size + x;
                    const h = height[i];
                    if (h > max) {
                        max = h;
                        visible[i] = true;
                    }
                }
            }

            // march up
            {
                var x: usize = n;
                var y: usize = size - 1;
                visible[y * size + x] = true;
                var max = height[y * size + x];
                y -= 1;
                while (y > 0) : (y -= 1) {
                    const i = y * size + x;
                    const h = height[i];
                    if (h > max) {
                        max = h;
                        visible[i] = true;
                    }
                }
            }

            // corners are vacuous
            {
                const last = size - 1;
                const lastOff = last * size;
                visible[0] = true;
                visible[last] = true;
                visible[lastOff] = true;
                visible[lastOff + last] = true;
            }
        }
    }
    try timing.markPhase(.computeVisibility);

    {
        try out.print("```visibility", .{});
        for (visible) |v, i| {
            const x = i % size;
            const sep = if (x == 0) "\n" else " ";
            const glyph = if (v) "+" else "-";
            try out.print("{s}{s}", .{ sep, glyph });
        }
        try out.print("\n```\n", .{});
    }
    try timing.markPhase(.displayVisibility);

    // how many trees are visible from outside the grid?
    {
        var sum: usize = 0;
        for (visible) |v| {
            if (v) sum += 1;
        }
        try out.print("\n# Part 1\n> {}\n", .{sum});
    }
    try timing.markPhase(.part1);

    // // FIXME part 2
    // {
    //     try out.print("\n# Part 2\n", .{});
    //     try out.print("> {}\n", .{42});
    //     try timing.markPhase(.part2);
    // }

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
