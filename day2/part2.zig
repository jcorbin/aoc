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
        return @intCast(u8, @enumToInt(self)) + 1;
    }

    pub fn vs(self: Move, other: Move) Outcome {
        return if (self.beats(other)) Outcome.win else if (other.beats(self)) Outcome.lose else Outcome.draw;
    }

    pub fn score(self: Move, other: Move) u8 {
        return self.worth() + self.vs(other).worth();
    }

    pub fn losesTo(self: Move) Move {
        return switch (self) {
            .rock => .paper,
            .paper => .scissors,
            .scissors => .rock,
        };
    }

    pub fn winsTo(self: Move) Move {
        return switch (self) {
            .rock => .scissors,
            .paper => .rock,
            .scissors => .paper,
        };
    }
};

const Outcome = enum(u2) {
    lose,
    draw,
    win,

    pub fn worth(self: Outcome) u8 {
        return @intCast(u8, @enumToInt(self)) * 3;
    }

    pub fn ensure(self: Outcome, other: Move) Move {
        return switch (self) {
            .lose => other.winsTo(),
            .draw => other,
            .win => other.losesTo(),
        };
    }
};

fn parseMove(s: []const u8, base: u8) ?Move {
    if (s.len != 1) return null;
    const c = s[0];
    return if (base <= c and c <= base + 2) @intToEnum(Move, c - base) else null;
}

fn parseOutcome(s: []const u8, base: u8) ?Outcome {
    if (s.len != 1) return null;
    const c = s[0];
    return if (base <= c and c <= base + 2) @intToEnum(Outcome, c - base) else null;
}

pub fn main() !void {
    const input = std.io.bufferedReader(std.io.getStdIn().reader()).reader();
    const output = std.io.getStdOut().writer();

    var score: u32 = 0;

    var buf = [_]u8{0} ** 4096;
    while (try input.readUntilDelimiterOrEof(buf[0..buf.len], '\n')) |line| {
        var tokens = std.mem.tokenize(u8, line, " ");
        const them = parseMove(tokens.next() orelse continue, 'A') orelse continue;
        const outcome = parseOutcome(tokens.next() orelse continue, 'X') orelse continue;
        const us = outcome.ensure(them);
        const round = us.score(them);
        score += round;
        try output.print("??? {any} (worth: {any} {any} ) <- {any} vs {any} <- {s}\n", .{ round, us.worth(), outcome, us, them, line });
        std.debug.assert(us.vs(them) == outcome);
    }

    try output.print("> {any}\n", .{score});
}
