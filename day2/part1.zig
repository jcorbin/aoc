const std = @import("std");

const Move = enum(u2) {
    rock,
    paper,
    scissors,

    pub fn beats(self: Move, other: Move) bool {
        return switch (self) {
            .rock => other == .scissors,
            .paper => other == .rock,
            .scissors => other == .paper,
        };
    }

    pub fn worth(self: Move) u8 {
        return @enumToInt(self) + 1;
    }

    pub fn score(self: Move, other: Move) u8 {
        if (self.beats(other)) return self.worth() + 6;
        if (!other.beats(self)) return self.worth() + 3;
        return self.worth();
    }
};

fn parseMove(s: []const u8, base: u8) ?Move {
    if (s.len != 1) return null;
    const c = s[0];
    return if (base <= c and c <= base + 2) @intToEnum(Move, c - base) else null;
}

pub fn main() !void {
    const input = std.io.bufferedReader(std.io.getStdIn().reader()).reader();
    const output = std.io.getStdOut().writer();

    var score: u32 = 0;

    var buf = [_]u8{0} ** 4096;
    while (try input.readUntilDelimiterOrEof(buf[0..buf.len], '\n')) |line| {
        var tokens = std.mem.tokenize(u8, line, " ");
        const them = parseMove(tokens.next() orelse continue, 'A') orelse continue;
        const us = parseMove(tokens.next() orelse continue, 'X') orelse continue;
        const round = us.score(them);
        score += round;
        try output.print("??? {any} (worth: {any} win: {any} lose: {any}) <- {any} vs {any} <- {s}\n", .{ round, us.worth(), us.beats(them), them.beats(us), us, them, line });
    }

    try output.print("> {any}\n", .{score});
}
