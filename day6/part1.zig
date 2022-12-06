const std = @import("std");

test "example" {
    const test_cases = [_]struct {
        data: []const u8,
        begin: usize,
    }{
        .{
            .data = "mjqjpqmgbljsphdztnvjfqwrcgsmlb",
            .begin = 7,
        },
        .{
            .data = "bvwbjplbgvbhsrlpgdmjqwftvncz",
            .begin = 5,
        },
        .{
            .data = "nppdvjthqldpwncqszvftbrmjlhg",
            .begin = 6,
        },
        .{
            .data = "nznrnfrfntjfmvfwmzdfjlvtqnbhcprsg",
            .begin = 10,
        },
        .{
            .data = "zcfzfwzzqfrljwzlrfnpqdbhtmscgvjw",
            .begin = 11,
        },
    };

    const allocator = std.testing.allocator;

    for (test_cases) |tc| {
        // TODO: can has sub-test?
        var buf = [_]u8{0} ** 1024;
        const expected = try std.fmt.bufPrint(buf[0..],
            \\> {d}
            \\
        , .{tc.begin});

        var input = std.io.fixedBufferStream(tc.data);
        var output = std.ArrayList(u8).init(allocator);
        defer output.deinit();

        run(&input, &output) catch |err| {
            std.debug.print("```pre-error output:\n{s}\n```\n", .{output.items});
            return err;
        };
        try std.testing.expectEqualStrings(expected, output.items);
    }
}

fn run(
    // TODO: better "any .reader()-able / any .writer()-able" interfacing
    input: anytype,
    output: anytype,
) !void {
    var in = input.reader();
    var out = output.writer();

    var offset: usize = 0;
    var buf = [_]u8{0} ** 4;
    const eobegin = while (true) : (offset += 1) {
        const next = in.readByte() catch |err| {
            if (err != error.EndOfStream) return err;
            break null;
        };
        buf[offset % 4] = next;
        if (offset < 3) continue;

        if (buf[0] == buf[1]) continue;
        if (buf[0] == buf[2]) continue;
        if (buf[0] == buf[3]) continue;
        if (buf[1] == buf[2]) continue;
        if (buf[1] == buf[3]) continue;
        if (buf[2] == buf[3]) continue;

        break offset + 1;
    } else null;

    try out.print("> {}\n", .{eobegin});
}

pub fn main() !void {
    // const allocator = std.heap.page_allocator;

    var input = std.io.getStdIn();
    var output = std.io.getStdOut();

    var bufin = std.io.bufferedReader(input.reader());
    var bufout = std.io.bufferedWriter(output.writer());

    try run(&bufin, &bufout);
    try bufout.flush();

    // TODO: argument parsing to steer input selection

    // TODO: sentinel-buffered output writer to flush lines progressively

    // TODO: input, output, and run-time metrics
}
