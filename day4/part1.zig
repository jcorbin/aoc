const std = @import("std");

test "example" {
    var input = std.io.fixedBufferStream(
        \\2-4,6-8
        \\2-3,4-5
        \\5-7,7-9
        \\2-8,3-7
        \\6-6,4-6
        \\2-6,4-8
        \\
    );

    var output = std.ArrayList(u8).init(std.testing.allocator);
    defer output.deinit();

    try run(input.reader(), output.writer());
    try std.testing.expectEqualStrings(
        \\- Range{ .lo = 3, .hi = 7 } ⊂ Range{ .lo = 2, .hi = 8 }
        \\- Range{ .lo = 6, .hi = 6 } ⊂ Range{ .lo = 4, .hi = 6 }
        \\> 2
        \\
    , output.items);

    // Some of the pairs have noticed that one of their assignments fully
    // contains the other. For example, `2-8` fully contains `3-7`, and `6-6`
    // is fully contained by `4-6`. In pairs where one assignment fully
    // contains the other, one Elf in the pair would be exclusively cleaning
    // sections their partner will already be cleaning, so these seem like the
    // most in need of reconsideration. In this example, there are *`2`* such
    // pairs.
}

const ParseRangeError = error{
    MissingDash,
};

const Range = struct {
    lo: u32,
    hi: u32,

    pub fn contains(self: @This(), other: @This()) bool {
        if (other.lo > self.hi) return false;
        if (other.hi < self.lo) return false;
        return other.lo >= self.lo and other.hi <= self.hi;
    }

    pub fn parse(str: []const u8) !@This() {
        const i = std.mem.indexOfScalar(u8, str, '-') orelse return ParseRangeError.MissingDash;
        return @This(){
            .lo = try std.fmt.parseInt(u32, str[0..i], 10),
            .hi = try std.fmt.parseInt(u32, str[i + 1 ..], 10),
        };
    }
};

const ParsePairError = error{
    MissingFirstToken,
    MissingSecondToken,
    ExtraTokens,
};

const Pair = struct {
    first: Range,
    second: Range,

    pub fn parse(str: []const u8) !@This() {
        var tokens = std.mem.tokenize(u8, str, ",");
        const firstToken = tokens.next() orelse return ParsePairError.MissingFirstToken;
        const secondToken = tokens.next() orelse return ParsePairError.MissingSecondToken;
        if (tokens.next() != null) return ParsePairError.ExtraTokens;

        return @This(){
            .first = try Range.parse(firstToken),
            .second = try Range.parse(secondToken),
        };
    }
};

// TODO: better "any reader / any writer" interfacing
fn run(input: anytype, output: anytype) !void {
    var total: u32 = 0;

    var buf = [_]u8{0} ** 4096;
    while (try input.readUntilDelimiterOrEof(buf[0..], '\n')) |line| {
        const pair = try Pair.parse(line);
        if (pair.second.contains(pair.first)) {
            total += 1;
            try output.print("- {} ⊂ {}\n", .{ pair.first, pair.second });
        } else if (pair.first.contains(pair.second)) {
            total += 1;
            try output.print("- {} ⊂ {}\n", .{ pair.second, pair.first });
        }
    }

    try output.print("> {d}\n", .{total});
}

pub fn main() !void {
    const input = std.io.bufferedReader(std.io.getStdIn().reader()).reader();
    const output = std.io.getStdOut().writer();
    try run(input, output);
}
