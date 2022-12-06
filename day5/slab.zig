const std = @import("std");
const Allocator = std.mem.Allocator;

/// Allocates n-chunks of T elements, which are then kept in an internal singly
/// linked list to be mass destroyed after all such elements are no longer needed.
/// - caller is responsible for any element reuse
/// - only aggregate destruction is supported; no per-element destruction
/// - usage pattern is similar to an arena, but type specific:
///
///     const Nodes = SlabChain(struct {
///         next: ?*@This() = null,
///         // node data fields to taste
///     }, 32).init(yourAllocator);
///     defer Nodes.deinit();
///
///     // build some graph of nodes or whatever
///     var nodeA = Nodes.create();
///     var nodeB = Nodes.create();
///     nodeA.next = nodeB;
pub fn SlabChain(comptime T: type, comptime n: usize) type {
    const Chunk = struct {
        free: usize = 0,
        chunk: [n]T = undefined,
        prior: ?*@This() = null,

        pub fn create(self: *@This()) ?*T {
            if (self.free >= self.chunk.len) return null;
            const node = &self.chunk[self.free];
            self.free += 1;
            return node;
        }
    };

    return struct {
        allocator: Allocator,
        last: ?*Chunk = null,

        pub fn init(allocator: Allocator) @This() {
            return .{ .allocator = allocator };
        }

        pub fn deinit(self: *@This()) void {
            var last = self.last;
            self.last = null;
            while (last) |slab| {
                last = slab.prior;
                self.allocator.destroy(slab);
            }
        }

        pub fn create(self: *@This()) !*T {
            while (true) {
                if (self.last) |last|
                    if (last.create()) |item|
                        return item;
                var new = try self.allocator.create(Chunk);
                new.* = .{ .prior = self.last };
                self.last = new;
            }
        }
    };
}
