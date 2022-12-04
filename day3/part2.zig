const std = @import("std");

test "example" {
    var input = std.io.fixedBufferStream(
        \\vJrwpWtwJgWrhcsFMMfFFhFp
        \\jqHRNqRjqzjGDLGLrsFMfFZSrLrFZsSL
        \\PmmdzqPrVvPwwTWBwg
        \\wMqvLMZHhHMvwLHjbvcjnnSBnvTQFn
        \\ttgJtRGJQctTZtZT
        \\CrZsJsPPZsGzwwsLwLmpwMDw
    );

    var output = std.ArrayList(u8).init(std.testing.allocator);
    defer output.deinit();

    try run(input.reader(), output.writer());
    try std.testing.expectEqualStrings(
        \\- r 18
        \\- Z 52
        \\> 70
        \\
    , output.items);

    // In the first group, the only item type that appears in all three
    // rucksacks is lowercase `r`; this must be their badges. In the second
    // group, their badge item type must be `Z`.
    //
    // Priorities for these items must still be found to organize the sticker
    // attachment efforts: here, they are `18` (`r`) for the first group and
    // `52` (`Z`) for the second group. The sum of these is *`70`*.
}

fn Sack(comptime CountType: type) type {
    return struct {
        const This = @This();

        items: [26 * 2]CountType,

        fn add(self: This, other: This) This {
            var sum = This{ .items = [_]CountType{0} ** 52 };
            for (self.items) |n, i|
                sum.items[i] = n + other.items[i];
            return sum;
        }
    };
}

fn char2prio(c: u8) ?u6 {
    const uc = c & 0x5f;
    if (uc < 'A' or uc > 'Z') return null;
    var i = (c & 0x5f) - 'A' + 1;
    if ((c & 0x20) == 0) i += 26;
    return @intCast(u6, i);
}

fn prio2char(p: u6) u8 {
    return if (p > 26) 'A' + (@intCast(u8, p) - 26 - 1) else 'a' + @intCast(u8, p) - 1;
}

const Sack3 = Sack(u3);

fn parseSackSet(in: []u8) Sack3 {
    var sack = Sack3{ .items = [_]u3{0} ** 52 };
    for (in) |c| {
        if (char2prio(c)) |prio| {
            sack.items[prio - 1] = 1;
        }
    }
    return sack;
}

const RunError = error{
    SpareBags,
    BadGroupCount,
    AmbiGroup,
    NoGroupBadge,
    DupeGroupBadge,
};

// TODO: how do we say "any Reader / any Writer"?
fn run(input: anytype, output: anytype) !void {
    var total: u64 = 0;

    var bagCount: u2 = 0;
    var group = Sack3{ .items = [_]u3{0} ** 52 };
    var used = Sack3{ .items = [_]u3{0} ** 52 };

    var buf = [_]u8{0} ** 4096;
    while (try input.readUntilDelimiterOrEof(buf[0..], '\n')) |line| {
        const bag = parseSackSet(line);

        group = group.add(bag);
        bagCount += 1;

        if ((bagCount % 3) == 0) {
            var have = false;
            for (group.items) |n, i| {
                if (n < 3) continue;
                if (n > 3) return RunError.BadGroupCount;
                if (have) return RunError.AmbiGroup;
                have = true;

                const prio = i + 1;
                const c = prio2char(@intCast(u6, prio));

                const wasUsed = used.items[i] != 0;

                used.items[i] = 1;
                total += prio;

                const note: []const u8 = if (wasUsed) " (reuse)" else "";
                try output.print("- {c} {d}{s}\n", .{ c, prio, note });
            }
            if (!have) return RunError.NoGroupBadge;

            group = Sack3{ .items = [_]u3{0} ** 52 };
            bagCount = 0;
        }
    }

    if (bagCount != 0) return RunError.SpareBags;

    try output.print("> {d}\n", .{total});
}

pub fn main() !void {
    const input = std.io.bufferedReader(std.io.getStdIn().reader()).reader();
    const output = std.io.getStdOut().writer();
    try run(input, output);
}
