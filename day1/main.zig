const std = @import("std");

pub fn main() !void {
    const input = std.io.bufferedReader(std.io.getStdIn().reader()).reader();
    const output = std.io.getStdOut().writer();

    // TODO how to abstract out a group-sum reader/iterator?

    var best: i32 = 0;

    while (true) {
        var any = false;
        var sum: i32 = 0;

        var buf = [_]u8{0} ** 4096;
        while (try input.readUntilDelimiterOrEof(buf[0..buf.len], '\n')) |line| {
            if (line.len == 0) {
                break;
            } else {
                const n = try std.fmt.parseInt(i32, line, 10);
                sum += n;
                any = true;
            }
        }

        if (!any) {
            break;
        }

        if (sum > best) {
            best = sum;
        }
    }

    try output.print("> {any}\n", .{best});
}
