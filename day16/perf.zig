const std = @import("std");
const Allocator = std.mem.Allocator;

pub fn Timing(comptime Tag: type) type {
    return struct {
        const Self = @This();

        const Data = std.ArrayList(struct {
            tag: Tag,
            time: u64,
        });

        arena: std.heap.ArenaAllocator,
        data: Data,
        overall: std.time.Timer,
        phase: std.time.Timer,

        pub fn start(allocator: Allocator) !Self {
            var overall = try std.time.Timer.start();
            return Self{
                .arena = std.heap.ArenaAllocator.init(allocator),
                .data = Data.init(allocator),
                .overall = overall,
                .phase = overall,
            };
        }

        pub fn deinit(self: *Self) void {
            self.data.deinit();
            self.arena.deinit();
        }

        pub const Timer = struct {
            self: *Self,
            tag: Tag,
            t: std.time.Timer,

            pub fn reset(tm: *@This()) void {
                tm.t.reset();
            }

            pub fn lap(tm: *@This()) !void {
                try tm.self.collect(tm.tag, tm.t.lap());
            }
        };

        pub fn timer(self: *Self, tag: Tag) !Timer {
            return Timer{
                .self = self,
                .tag = tag,
                .t = try std.time.Timer.start(),
            };
        }

        pub fn collect(self: *Self, tag: Tag, time: u64) !void {
            try self.data.append(.{
                .tag = tag,
                .time = time,
            });
        }

        pub fn markPhase(self: *Self, tag: Tag) !void {
            try self.collect(tag, self.phase.lap());
        }

        pub fn finish(self: *Self, tag: Tag) !void {
            const t = self.overall.lap();
            self.phase = self.overall;
            try self.collect(tag, t);
        }

        pub fn printDebugReport(self: *Self) void {
            std.debug.print("# Timing\n\n", .{});
            for (self.data.items) |item| {
                std.debug.print("- {} {}\n", .{ item.time, item.tag });
            }
            std.debug.print("\n", .{});
        }
    };
}
