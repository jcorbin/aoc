const std = @import("std");

test "example" {
    const test_cases = [_]struct {
        data: []const u8,
        packetStart: usize,
        messageStart: usize,
    }{
        .{
            .data = "mjqjpqmgbljsphdztnvjfqwrcgsmlb",
            .packetStart = 7,
            .messageStart = 19,
        },
        .{
            .data = "bvwbjplbgvbhsrlpgdmjqwftvncz",
            .packetStart = 5,
            .messageStart = 23,
        },
        .{
            .data = "nppdvjthqldpwncqszvftbrmjlhg",
            .packetStart = 6,
            .messageStart = 23,
        },
        .{
            .data = "nznrnfrfntjfmvfwmzdfjlvtqnbhcprsg",
            .packetStart = 10,
            .messageStart = 29,
        },
        .{
            .data = "zcfzfwzzqfrljwzlrfnpqdbhtmscgvjw",
            .packetStart = 11,
            .messageStart = 26,
        },
    };

    const allocator = std.testing.allocator;

    for (test_cases) |tc| {
        // TODO: can has sub-test?
        var buf = [_]u8{0} ** 1024;
        const expected = try std.fmt.bufPrint(buf[0..],
            \\> {d} {d}
            \\
        , .{ tc.packetStart, tc.messageStart });

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

fn QGram(comptime Q: usize) type {
    return struct {
        data: [Q]u8 = [_]u8{0} ** Q,

        fn anySame(self: @This()) bool {
            for (self.data) |a, i| {
                for (self.data[i + 1 ..]) |b| {
                    if (a == b) return true;
                }
            }

            // TODO why not?
            // var i: usize = 0;
            // inline while (i < Q) : (i += 1) {
            //     const a = self.data[i];
            //     var j: usize = i + 1;
            //     inline while (j < Q) : (j += 1) {
            //         if (a == self.data[j]) return true;
            //     }
            // }

            return false;
        }
    };
}

fn run(
    // TODO: better "any .reader()-able / any .writer()-able" interfacing
    input: anytype,
    output: anytype,
) !void {
    var in = input.reader();
    var out = output.writer();

    var packetStart = QGram(4){};
    var messageStart = QGram(14){};

    var offset: usize = 0;

    const eoPacketStart = while (true) {
        const next = in.readByte() catch |err| {
            if (err != error.EndOfStream) return err;
            break null;
        };

        packetStart.data[offset % packetStart.data.len] = next;
        messageStart.data[offset % messageStart.data.len] = next;
        offset += 1;

        if (offset < packetStart.data.len) continue;
        if (packetStart.anySame()) continue;

        break offset;
    } else null;

    const eoMessageStart = while (true) {
        const next = in.readByte() catch |err| {
            if (err != error.EndOfStream) return err;
            break null;
        };

        messageStart.data[offset % messageStart.data.len] = next;
        offset += 1;

        if (offset < messageStart.data.len) continue;
        if (messageStart.anySame()) continue;

        break offset;
    } else null;

    try out.print("> {} {}\n", .{ eoPacketStart, eoMessageStart });
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
