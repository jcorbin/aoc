const std = @import("std");

fn greaterThan(_: void, a: i32, b: i32) std.math.Order {
    return std.math.order(a, b);
}

pub fn main() !void {
    const input = std.io.bufferedReader(std.io.getStdIn().reader()).reader();
    const output = std.io.getStdOut().writer();

    // TODO how to abstract out a group-sum reader/iterator?

    var top = std.PriorityQueue(i32, void, greaterThan)
        .init(std.heap.page_allocator, undefined);
    defer top.deinit();

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

        if (top.count() < 3) {
            try top.add(sum);
        } else if (top.peek()) |best| {
            if (sum > best) {
                try top.update(best, sum);
            }
        }
    }

    var sum: i32 = 0;
    while (top.removeOrNull()) |best| {
        sum += best;
        try output.print("... {any}\n", .{best});
    }
    try output.print("> {any}\n", .{sum});
}
