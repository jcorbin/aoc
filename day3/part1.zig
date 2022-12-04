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
        \\- p 16
        \\- L 38
        \\- P 42
        \\- v 22
        \\- t 20
        \\- s 19
        \\> 157
        \\
    , output.items);

    // - The first rucksack contains the items `vJrwpWtwJgWrhcsFMMfFFhFp`,
    //   which means its first compartment contains the items `vJrwpWtwJgWr`,
    //   while the second compartment contains the items `hcsFMMfFFhFp`. The
    //   only item type that appears in both compartments is lowercase `p`.
    // - The second rucksack's compartments contain `jqHRNqRjqzjGDLGL` and
    //   `rsFMfFZSrLrFZsSL`. The only item type that appears in both
    //   compartments is uppercase `L`.
    // - The third rucksack's compartments contain `PmmdzqPrV` and `vPwwTWBwg`;
    //   the only common item type is uppercase `P`.
    // - The fourth rucksack's compartments only share item type `v`.
    // - The fifth rucksack's compartments only share item type `t`.
    // - The sixth rucksack's compartments only share item type `s`.
    //
    // In the above example, the priority of the item type that appears in both
    // compartments of each rucksack is `16` (`p`), `38` (`L`), `42` (`P`),
    // `22` (`v`), `20` (`t`), and `19` (`s`); the sum of these is *`157`*.
}

const Sack = struct {
    items: [26 * 2]u2,

    fn add(self: Sack, other: Sack) Sack {
        var sum = Sack{ .items = [_]u2{0} ** 52 };
        for (self.items) |n, i|
            sum.items[i] = n + other.items[i];
        return sum;
    }
};

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

fn parseSackSet(in: []u8) Sack {
    var sack = Sack{ .items = [_]u2{0} ** 52 };
    for (in) |c| {
        if (char2prio(c)) |prio| {
            sack.items[prio - 1] = 1;
        }
    }
    return sack;
}

// TODO: how do we say "any Reader / any Writer"?
fn run(input: anytype, output: anytype) !void {
    var total: u64 = 0;

    var buf = [_]u8{0} ** 4096;
    while (try input.readUntilDelimiterOrEof(buf[0..], '\n')) |line| {
        const mid = line.len / 2;
        const one = parseSackSet(line[0..mid]);
        const two = parseSackSet(line[mid..]);
        const sum = one.add(two);

        // TODO: would prefer to have this be an iterator
        var any = false;
        for (sum.items) |n, i| {
            if (n > 1) {
                any = true;
                const prio = i + 1;
                const c = prio2char(@intCast(u6, prio));
                total += prio;
                try output.print("- {c} {d}\n", .{ c, prio });
            }
        }
        if (!any) {
            try output.print("- NONE\n", .{});
        }
    }

    try output.print("> {d}\n", .{total});
}

pub fn main() !void {
    const input = std.io.bufferedReader(std.io.getStdIn().reader()).reader();
    const output = std.io.getStdOut().writer();
    try run(input, output);
}
