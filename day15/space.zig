const std = @import("std");
const math = std.math;
const assert = std.debug.assert;

pub fn Space(comptime Int: type) type {
    const int_info = @typeInfo(Int);

    const UInt = switch (int_info) {
        .ComptimeInt => comptime_int,
        .Int => |int_info| std.meta.Int(.unsigned, int_info.bits),
        else => @compileError("Space only accepts integers"),
    };

    return struct {
        pub const Point = std.meta.Vector(2, Int);
        pub const UPoint = std.meta.Vector(2, UInt);

        const LineIterator = struct {
            from: Point,
            to: Point,
            d: Point,
            i: usize = 0,
            len: usize,

            const Self = @This();

            pub fn next(self: *Self) ?Point {
                if (self.i >= self.len) return null;
                const r = self.from;
                self.from = r + self.d;
                self.i += 1;
                return r;
            }
        };

        pub fn line(from: Point, to: Point) LineIterator {
            const delta = to - from;
            return .{
                .from = from,
                .to = to,
                .d = math.sign(delta),
                .len = @intCast(usize, @maximum(
                    math.absInt(delta[0]) catch unreachable,
                    math.absInt(delta[1]) catch unreachable,
                )),
            };
        }

        pub const Rect = struct {
            from: Point,
            to: Point,

            const Self = @This();

            pub fn width(self: Self) UInt {
                return if (self.to[0] > self.from[0]) @intCast(UInt, self.to[0] - self.from[0]) else 0;
            }

            pub fn height(self: Self) UInt {
                return if (self.to[1] > self.from[1]) @intCast(UInt, self.to[1] - self.from[1]) else 0;
            }

            pub fn size(self: Self) UPoint {
                return .{ self.width(), self.height() };
            }

            pub fn relativize(self: Self, p: Point) UPoint {
                assert(self.contains(p));
                const d = p - self.from;
                return .{
                    @intCast(UInt, d[0]),
                    @intCast(UInt, d[1]),
                };
            }

            pub fn contains(self: Self, to: Point) bool {
                return to[0] >= self.from[0] and
                    to[1] >= self.from[1] and
                    to[0] < self.to[0] and
                    to[1] < self.to[1];
            }

            pub fn expandTo(self: *Self, to: Point) void {
                const top = Point{ to[0] + 1, to[1] + 1 };
                self.* = if (self.width() == 0 or self.height() == 0) .{
                    .from = to,
                    .to = top,
                } else .{
                    .from = @minimum(self.from, to),
                    .to = @maximum(self.to, top),
                };
            }
        };
    };
}
