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
        \\- Range{ .lo = 7, .hi = 7 } = Range{ .lo = 5, .hi = 7 } ∩ Range{ .lo = 7, .hi = 9 }
        \\- Range{ .lo = 3, .hi = 7 } = Range{ .lo = 2, .hi = 8 } ∩ Range{ .lo = 3, .hi = 7 }
        \\- Range{ .lo = 6, .hi = 6 } = Range{ .lo = 6, .hi = 6 } ∩ Range{ .lo = 4, .hi = 6 }
        \\- Range{ .lo = 4, .hi = 6 } = Range{ .lo = 2, .hi = 6 } ∩ Range{ .lo = 4, .hi = 8 }
        \\> 4
        \\
    , output.items);

    // In the above example, the first two pairs (`2-4,6-8` and `2-3,4-5`)
    // don't overlap, while the remaining four pairs (`5-7,7-9`, `2-8,3-7`,
    // `6-6,4-6`, and `2-6,4-8`) do overlap:
    // - `5-7,7-9` overlaps in a single section, `7`.
    // - `2-8,3-7` overlaps all of the sections `3` through `7`.
    // - `6-6,4-6` overlaps in a single section, `6`.
    // - `2-6,4-8` overlaps in sections `4`, `5`, and `6`.
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

    pub fn size(self: @This()) u32 {
        return if (self.hi >= self.lo) 1 + self.hi - self.lo else 0;
    }

    pub fn overlap(self: @This(), other: @This()) @This() {
        return .{
            .lo = @maximum(self.lo, other.lo),
            .hi = @minimum(self.hi, other.hi),
        };
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
        const overlap = pair.first.overlap(pair.second);
        if (overlap.size() > 0) {
            total += 1;
            try output.print("- {} = {} ∩ {}\n", .{ overlap, pair.first, pair.second });
        }
    }

    try output.print("> {d}\n", .{total});
}

pub fn main() !void {
    const input = std.io.bufferedReader(std.io.getStdIn().reader()).reader();
    const output = std.io.getStdOut().writer();
    try run(input, output);
}
