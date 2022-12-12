const std = @import("std");
const isDigit = std.ascii.isDigit;

pub fn DelimScanner(
    comptime Reader: anytype,
    comptime delim: u8,
    comptime bufferSize: usize,
) type {
    return struct {
        buffer: [bufferSize]u8 = [_]u8{0} ** bufferSize,
        reader: Reader,
        cur: Cursor = .{ .buf = "" },

        const Self = @This();

        pub fn next(self: *Self) !?*Cursor {
            var line = (try self.reader.readUntilDelimiterOrEof(self.buffer[0..], delim)) orelse return null;
            self.cur.count += 1;
            self.cur.buf = line;
            self.cur.i = 0;
            return &self.cur;
        }
    };
}

pub fn lineScanner(reader: anytype) DelimScanner(@TypeOf(reader), '\n', 4096) {
    return .{ .reader = reader };
}

pub const Cursor = struct {
    buf: []const u8,
    i: usize = 0,
    count: usize = 0,

    const Self = @This();

    pub fn reset(self: *Self) void {
        self.i = 0;
    }

    pub fn live(self: Self) bool {
        return self.i < self.buf.len;
    }

    pub fn rem(self: Self) ?[]const u8 {
        return if (self.i < self.buf.len) self.buf[self.i..] else null;
    }

    pub fn peek(self: Self) ?u8 {
        return if (self.i < self.buf.len) self.buf[self.i] else null;
    }

    pub fn consume(self: *Self) ?u8 {
        if (self.peek()) |c| {
            self.i += 1;
            return c;
        }
        return null;
    }

    pub fn have(self: *Self, wanted: u8) bool {
        const c = self.peek() orelse return false;
        if (c == wanted) {
            self.i += 1;
            return true;
        }
        return false;
    }

    pub fn expectOrEnd(self: *Self, wanted: u8, err: anyerror) !void {
        if (!self.have(wanted) and self.live()) return err;
    }

    pub fn expectEnd(self: *Self, err: anyerror) !void {
        if (self.live()) return err;
    }

    pub fn expect(self: *Self, wanted: u8, err: anyerror) !void {
        if (!self.have(wanted)) return err;
    }

    pub fn expectNM(self: *Self, wanted: u8, atleast: usize, upto: usize, err: anyerror) !void {
        var got: usize = 0;
        while (got < upto) : (got += 1)
            if (!self.have(wanted))
                if (got < atleast) return err else return;
    }

    pub fn expectN(self: *Self, wanted: u8, n: usize, err: anyerror) !void {
        return self.expectNM(wanted, n, n, err);
    }

    pub fn expectStar(self: *Self, wanted: u8) void {
        self.expectNM(wanted, 0, 2 + self.buf.len - self.i, error.Nope) catch unreachable;
    }

    pub fn expectPlus(self: *Self, wanted: u8, err: anyerror) !void {
        return self.expectNM(wanted, 1, 2 + self.buf.len - self.i, err);
    }

    pub fn expectStr(self: *Self, wanted: []const u8, err: anyerror) !void {
        for (wanted) |c|
            try self.expect(c, err);
    }

    fn peekToken(self: *Self) ?[]const u8 {
        var i = self.i;
        var j = i;
        next: while (j < self.buf.len) : (j += 1) {
            switch (self.buf[j]) {
                // TODO user provided include/exclude sets
                ' ', '\t' => break :next,
                else => continue,
            }
        }
        return if (j > i) self.buf[i..j] else null;
    }

    pub fn haveLiteral(self: *Self, expected: []const u8) bool {
        var i = self.i;
        for (expected) |c| {
            if (i >= self.buf.len) return false;
            if (self.buf[i] != c) return false;
            i += 1;
        }
        self.i = i;
        return true;
    }

    pub fn expectLiteral(self: *Self, expected: []const u8, err: anyerror) !void {
        if (!self.haveLiteral(expected)) return err;
    }

    pub fn consumeToken(self: *Self) ?[]const u8 {
        const token = self.peekToken() orelse return null;
        self.i += token.len;
        return token;
    }

    pub fn expectToken(self: *Self, err: anyerror) ![]const u8 {
        return self.consumeToken() orelse err;
    }

    pub fn consumeInt(self: *Self, comptime T: type, radix: u8) ?T {
        var i = self.i;
        var j = i;
        while (j < self.buf.len) switch (self.buf[j]) {
            '-', '+' => {
                if (j > i) break;
                j += 1;
            },
            '_', '0'...'9', 'A'...'Z', 'a'...'z' => j += 1,
            else => break,
        };
        if (j == i) return null;
        const token = self.buf[self.i..j];
        const n = std.fmt.parseInt(T, token, radix) catch return null;
        self.i = j;
        return n;
    }

    pub fn expectInt(self: *Self, comptime T: type, radix: u8, err: anyerror) !T {
        return self.consumeInt(T, radix) orelse err;
    }
};
