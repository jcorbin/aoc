const std = @import("std");
const fmt = std.fmt;
const mem = std.mem;

const Allocator = mem.Allocator;

pub const Grid = struct {
    allocator: Allocator,
    width: usize,
    lineOffset: usize,
    lineStride: usize,
    buf: []u8,

    const Self = @This();

    pub fn init(allocator: Allocator, opts: struct {
        width: usize,
        height: usize,
        linePrefix: []const u8 = "",
        lineSuffix: []const u8 = "\n",
        fill: u8 = ' ',
    }) !Self {
        const lineStride = opts.linePrefix.len + opts.width + opts.lineSuffix.len;
        const memSize = lineStride * opts.height;
        var buf = try allocator.alloc(u8, memSize);
        mem.set(u8, buf, opts.fill);

        if (opts.linePrefix.len > 0) {
            var i: usize = 0;
            while (i < buf.len) : (i += lineStride)
                mem.copy(u8, buf[i..], opts.linePrefix);
        }

        if (opts.lineSuffix.len > 0) {
            var i = opts.linePrefix.len + opts.width;
            while (i < buf.len) : (i += lineStride)
                mem.copy(u8, buf[i..], opts.lineSuffix);
        }

        return Self{
            .allocator = allocator,
            .width = opts.width,
            .lineOffset = opts.linePrefix.len,
            .lineStride = lineStride,
            .buf = buf,
        };
    }

    pub fn deinit(self: *Self) void {
        self.allocator.free(self.buf);
    }

    pub fn format(self: Self, comptime _: []const u8, _: fmt.FormatOptions, writer: anytype) !void {
        try writer.print("{s}", .{mem.trimRight(u8, self.buf, "\n")});
    }

    pub fn ref(self: Self, x: usize, y: usize) *u8 {
        const i = self.lineOffset + y * self.lineStride + x;
        return &self.buf[i];
    }

    pub fn set(self: Self, x: usize, y: usize, c: u8) void {
        self.ref(x, y).* = c;
    }

    pub fn get(self: Self, x: usize, y: usize) u8 {
        return self.ref(x, y).*;
    }
};
