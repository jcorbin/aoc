const std = @import("std");
const Allocator = std.mem.Allocator;

test "example" {
    const example =
        \\such data
        \\
    ;

    const expected =
        \\> 42
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

fn run(
    under_allocator: Allocator,

    // TODO: better "any .reader()-able / any .writer()-able" interfacing
    input: anytype,
    output: anytype,
) !void {
    var arena = std.heap.ArenaAllocator.init(under_allocator);
    defer arena.deinit();

    // FIXME: maybe use this
    // const allocator = arena.allocator();

    var buf = [_]u8{0} ** 4096;
    var in = input.reader();
    var out = output.writer();

    // FIXME: such computer
    while (try in.readUntilDelimiterOrEof(buf[0..], '\n')) |line| {
        // FIXME: much line
        _ = line;
    }

    // FIXME: very answer
    try out.print("> {}\n", .{42});
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
