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

        pub fn init(allocator: Allocator) Self {
            return .{
                .arena = std.heap.ArenaAllocator.init(allocator),
                .data = Data.init(allocator),
            };
        }

        pub fn deinit(self: *Self) void {
            self.arena.deinit();
            self.data.deinit();
        }

        pub fn collect(self: *Self, tag: Tag, time: u64) !void {
            try self.data.append(.{
                .tag = tag,
                .time = time,
            });
        }
    };
}
