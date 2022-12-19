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

    pub fn carp(self: *Self, err: anyerror) anyerror {
        const space = " " ** 4096;
        std.debug.print(
            \\Unable to parse line #{}:
            \\> {s}
            \\  {s}^-- {} here
            \\
        , .{
            self.count,
            self.buf,
            space[0..self.i],
            err,
        });
        return err;
    }

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

    pub const Wanted = union(enum) {
        just: u8,
        range: struct { min: u8, max: u8 },
        any: []const u8,
        literal: []const u8,

        pub fn have(wanted: @This(), s: []const u8) ?usize {
            switch (wanted) {
                .literal => |lit| if (std.mem.startsWith(u8, s, lit)) return lit.len,
                .just => |w| if (s.len > 0 and s[0] == w) return 1,
                .range => |r| if (s.len > 0 and r.min <= s[0] and s[0] <= r.max) return 1,
                .any => |w| if (s.len > 0 and std.mem.indexOfScalar(u8, w, s[0]) != null) return 1,
            }
            return null;
        }
    };

    pub fn have(self: *Self, comptime wanted: Wanted) ?[]const u8 {
        const start = self.i;
        if (wanted.have(self.buf[self.i..])) |n| {
            self.i += n;
            return self.buf[start .. start + n];
        }
        return null;
    }

    pub fn haveNM(self: *Self, comptime atleast: usize, comptime upto: usize, comptime wanted: Wanted) ?[]const u8 {
        var got: usize = 0;
        const i = self.i;
        var j = i;
        while (got < upto and j < self.buf.len) : (got += 1) {
            const n = wanted.have(self.buf[j..]) orelse break;
            j += n;
        }
        if (got >= atleast) {
            self.i = j;
            return self.buf[i..j];
        }
        return null;
    }

    pub fn haveN(self: *Self, comptime n: usize, comptime wanted: Wanted) ?[]const u8 {
        return self.haveNM(n, n, wanted);
    }

    pub fn star(self: *Self, comptime wanted: Wanted) []const u8 {
        const i = self.i;
        var j = i;
        while (j < self.buf.len) {
            const n = wanted.have(self.buf[j..]) orelse break;
            j += n;
        }
        self.i = j;
        return self.buf[i..j];
    }

    pub fn plus(self: *Self, comptime wanted: Wanted) ?[]const u8 {
        const i = self.i;
        var j = i;
        while (j < self.buf.len) {
            const n = wanted.have(self.buf[j..]) orelse break;
            j += n;
        }
        if (j > i) {
            self.i = j;
            return self.buf[i..j];
        }
        return null;
    }
};
