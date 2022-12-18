const std = @import("std");

const math = std.math;
const mem = std.mem;

const Allocator = mem.Allocator;
const PriorityQueue = std.PriorityQueue;

pub const Action = enum {
    halt,
    take,
    skip,
    queue,
};

pub fn Queue(
    comptime State: type,
    comptime StateIterator: type,
    comptime Context: type,
    comptime compareFn: fn (context: Context, a: State, b: State) math.Order,
    comptime expandFn: fn (context: Context, state: State) StateIterator,
    comptime consumeFn: fn (context: Context, state: State) Action,
    comptime destroyFn: fn (context: Context, state: State) void,
) type {
    // TODO assert that StateIterator has a next() !?State method

    return struct {
        const Self = @This();

        const PQ = PriorityQueue(State, Context, compareFn);

        context: Context,
        queue: PQ,
        halted: bool = false,

        pub fn init(allocator: Allocator, context: Context) Self {
            return Self{
                .context = context,
                .queue = PQ.init(allocator, context),
            };
        }

        pub fn deinit(self: *Self) void {
            for (self.queue.items[0..self.queue.len]) |state|
                destroyFn(self.context, state);
            self.queue.deinit();
        }

        pub fn add(self: *Self, state: State) !void {
            switch (consumeFn(self.context, state)) {
                .halt => self.halted = true,
                .take => {},
                .skip => destroyFn(self.context, state),
                .queue => try self.queue.add(state),
            }
        }

        pub fn done(self: *Self) bool {
            return self.halted or self.queue.len == 0;
        }

        pub fn run(self: *Self) !void {
            while (!self.halted)
                if (!try self.expand())
                    break;
        }

        pub fn runUpto(self: *Self, max_rounds: usize) !usize {
            var rounds: usize = 0;
            while (!self.halted) {
                if (!try self.expand()) break;
                rounds += 1;
                if (rounds >= max_rounds) break;
            }
            return rounds;
        }

        pub fn expand(self: *Self) !bool {
            if (self.halted) return false;
            var state = self.queue.removeOrNull() orelse return false;
            var it = expandFn(self.context, state);
            defer it.deinit();
            while (try it.next()) |next| {
                try self.add(next);
                if (self.halted) break;
            }
            return true;
        }
    };
}
