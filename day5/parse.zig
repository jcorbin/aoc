const std = @import("std");
const isDigit = std.ascii.isDigit;

pub const Cursor = struct {
    buf: []const u8,
    i: usize,

    pub fn make(buf: []const u8) @This() {
        return .{ .buf = buf, .i = 0 };
    }

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

    pub fn expectn(self: *@This(), wanted: u8, count: usize, err: anyerror) !void {
        var n: usize = 0;
        while (n < count) : (n += 1)
            try self.expect(wanted, err);
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
