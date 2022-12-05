const std = @import("std");

test "example" {
    const example =
        \\such data
        \\
    ;

    const expected =
        \\> 42
        \\
    ;

    var input = std.io.fixedBufferStream(example);
    var output = std.ArrayList(u8).init(std.testing.allocator);
    defer output.deinit();

    try run(&input, &output);
    try std.testing.expectEqualStrings(expected, output.items);
}

// TODO: better "any .reader()-able / any .writer()-able" interfacing
fn run(input: anytype, output: anytype) !void {
    var in = input.reader();
    var out = output.writer();

    // TODO: much computer
    _ = in;

    try out.print("> 42\n", .{}); // FIXME: very answer
}

pub fn main() !void {
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
