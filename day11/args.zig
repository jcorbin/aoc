const std = @import("std");
const Allocator = std.mem.Allocator;

// TODO: higher level type wrapped around Parser that provides more integrated
//       usage error reporting and help display

pub const Parser = struct {
    // allocator: Allocator,
    arena: std.heap.ArenaAllocator,
    it: std.process.ArgIterator,

    i: usize = 0,
    progArg: [:0]const u8 = "",

    const Self = @This();

    pub fn init(allocator: Allocator) !Self {
        var arena = std.heap.ArenaAllocator.init(allocator);
        var it = try std.process.argsWithAllocator(arena.allocator());
        const progArg = try (it.next(arena.allocator()) orelse return error.MissingArg0);
        return Self{
            // .allocator = allocator,
            .arena = arena,
            .it = it,
            .progArg = progArg,
        };
    }

    pub fn deinit(self: *Self) void {
        self.arena.deinit();
        // self.allocator.free(self.arena);
    }

    pub fn progName(self: Self) []const u8 {
        if (std.mem.lastIndexOf(u8, self.progArg, "/")) |i| {
            return self.progArg[i + 1 ..];
        } else {
            return self.progArg;
        }
    }

    pub fn next(self: *Self) !?Arg {
        var have = try (self.it.next(self.arena.allocator()) orelse return null);
        const index = self.i;
        self.i += 1;
        return Arg{ .index = index, .have = have };
    }
};

const Arg = struct {
    index: usize,
    have: [:0]const u8,

    const Self = @This();

    pub fn is(self: Self, anyOf: anytype) bool {
        const AnyOfType = @TypeOf(anyOf);
        // XXX std.meta.trait.isTuple(AnyOfType) doesn't seem to work...
        const any_of_info = @typeInfo(AnyOfType);
        if (any_of_info != .Struct) {
            @compileError("Expected tuple or struct argument, found " ++ @typeName(AnyOfType));
        }
        inline for (anyOf) |expected|
            if (std.mem.eql(u8, self.have, expected))
                return true;
        return false;
    }

    pub fn parseInt(self: Self, comptime T: type, radix: u8) !T {
        return std.fmt.parseInt(T, self.have, radix);
    }
};
