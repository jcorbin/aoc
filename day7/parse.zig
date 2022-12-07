const std = @import("std");
const isDigit = std.ascii.isDigit;

pub fn DelimScanner(
    comptime Reader: anytype,
    comptime delim: u8,
    comptime bufferSize: usize,
) type {
    return struct {
        const delim = delim;

        buffer: [bufferSize]u8 = [_]u8{0} ** bufferSize,
        reader: Reader,
        count: usize = 0,

        const Self = @This();

        pub fn next(self: *Self) !?Cursor {
            var line = (try self.reader.readUntilDelimiterOrEof(self.buffer[0..], delim)) orelse return null;
            self.count += 1;
            return Cursor{
                .buf = line,
                .count = self.count,
            };
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

    pub fn live(self: @This()) bool {
        return self.i < self.buf.len;
    }

    pub fn rem(self: @This()) ?[]const u8 {
        return if (self.i < self.buf.len) self.buf[self.i..] else null;
    }

    pub fn peek(self: @This()) ?u8 {
        return if (self.i < self.buf.len) self.buf[self.i] else null;
    }

    pub fn consume(self: *@This()) ?u8 {
        if (self.peek()) |c| {
            self.i += 1;
            return c;
        }
        return null;
    }

    pub fn have(self: *@This(), wanted: u8) bool {
        const c = self.peek() orelse return false;
        if (c == wanted) {
            self.i += 1;
            return true;
        }
        return false;
    }

    pub fn expectOrEnd(self: *@This(), wanted: u8, err: anyerror) !void {
        if (!self.have(wanted) and self.live()) return err;
    }

    pub fn expectEnd(self: *@This(), err: anyerror) !void {
        if (self.live()) return err;
    }

    pub fn expect(self: *@This(), wanted: u8, err: anyerror) !void {
        if (!self.have(wanted)) return err;
    }

    pub fn expectNM(self: *@This(), wanted: u8, atleast: usize, upto: usize, err: anyerror) !void {
        var got: usize = 0;
        while (got < upto) : (got += 1)
            if (!self.have(wanted))
                if (got < atleast) return err else return;
    }

    pub fn expectN(self: *@This(), wanted: u8, n: usize, err: anyerror) !void {
        return self.expectNM(wanted, n, n, err);
    }

    pub fn expectStar(self: *@This(), wanted: u8) void {
        self.expectNM(wanted, 0, 2 + self.buf.len - self.i, error.Nope) catch unreachable;
    }

    pub fn expectPlus(self: *@This(), wanted: u8, err: anyerror) !void {
        return self.expectNM(wanted, 1, 2 + self.buf.len - self.i, err);
    }

    pub fn expectStr(self: *@This(), wanted: []const u8, err: anyerror) !void {
        for (wanted) |c|
            try self.expect(c, err);
    }

    pub fn readInt(self: *@This(), comptime T: type) ?T {
        const base = 10; // TODO parameterize?
        var n: T = 0;
        var any = false;
        while (self.peek()) |c| {
            if (!isDigit(c)) break;
            any = true;
            n = n * base + @intCast(T, c - '0');
            self.i += 1;
        }
        return if (any) n else null;
    }

    pub fn expectInt(self: *@This(), comptime T: type, err: anyerror) !T {
        return self.readInt(T) orelse err;
    }
};
