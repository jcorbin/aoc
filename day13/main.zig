const std = @import("std");
const Allocator = std.mem.Allocator;
const math = std.math;

test "example" {
    const test_cases = [_]struct {
        input: []const u8,
        expected: []const u8,
        config: Config,
    }{
        // Part 1 example
        .{
            .config = .{
                .verbose = 1,
            },
            .input = 
            \\[1,1,3,1,1]
            \\[1,1,5,1,1]
            \\
            \\[[1],[2,3,4]]
            \\[[1],4]
            \\
            \\[9]
            \\[[8,7,6]]
            \\
            \\[[4,4],4,4]
            \\[[4,4],4,4,4]
            \\
            \\[7,7,7,7]
            \\[7,7,7]
            \\
            \\[]
            \\[3]
            \\
            \\[[[]]]
            \\[[]]
            \\
            \\[1,[2,[3,[4,[5,6,7]]]],8,9]
            \\[1,[2,[3,[4,[5,6,0]]]],8,9]
            \\
            ,
            .expected = 
            \\# Solution
            \\1. correct
            \\2. correct
            \\4. correct
            \\6. correct
            \\> 13
            \\
            ,
        },
    };

    const allocator = std.testing.allocator;

    for (test_cases) |tc, i| {
        std.debug.print(
            \\
            \\Test Case {}
            \\===
            \\
        , .{i});
        var input = std.io.fixedBufferStream(tc.input);
        var output = std.ArrayList(u8).init(allocator);
        defer output.deinit();
        run(allocator, &input, &output, tc.config) catch |err| {
            std.debug.print("```pre-error output:\n{s}\n```\n", .{output.items});
            return err;
        };
        try std.testing.expectEqualStrings(tc.expected, output.items);
    }
}

const Parse = @import("parse.zig");
const Timing = @import("perf.zig").Timing(enum {
    parse,
    parseLine,
    compare,
    compareOne,
    count,
    report,
    overall,
});

const Config = struct {
    verbose: usize = 0,
};

const Number = usize;
const List = std.ArrayListUnmanaged(Data);

const Data = union(enum) {
    number: Number,
    list: List,

    pub fn order(a: Data, b: Data) math.Order {
        var ait = a.datums();
        var bit = b.datums();
        var res = math.Order.eq;
        while (true) {
            const an = ait.next();
            const bn = bit.next();
            if (an == null and bn == null) return res;
            if (res != .eq) continue;

            const ad = an orelse return .lt;
            const bd = bn orelse return .gt;

            res = if (ad == .number and bd == .number)
                std.math.order(ad.number, bd.number)
            else
                Data.order(ad, bd);
        }
    }

    pub fn format(value: Data, comptime _: []const u8, _: std.fmt.FormatOptions, writer: anytype) !void {
        switch (value) {
            .number => |n| return writer.print("{d}", .{n}),
            .list => |l| {
                try writer.print("[", .{});
                for (l.items) |item, i|
                    if (i == 0)
                        try writer.print("{}", .{item})
                    else
                        try writer.print(", {}", .{item});
                try writer.print("]", .{});
            },
        }
    }

    const Datums = struct {
        data: ?Data,

        pub fn next(it: *Datums) ?Data {
            switch (it.data orelse return null) {
                .number => |n| {
                    it.data = null;
                    return Data{ .number = n };
                },
                .list => |l| if (l.items.len > 0) {
                    it.data = Data{ .list = .{ .items = l.items[1..] } };
                    return l.items[0];
                } else return null,
            }
        }
    };

    pub fn datums(self: Data) Datums {
        return .{ .data = self };
    }
};

const Pair = struct {
    left: Data,
    right: Data,
};

const Builder = struct {
    allocator: Allocator,

    left: ?Data = null,
    right: ?Data = null,
    pairs: std.ArrayListUnmanaged(Pair) = .{},

    const Self = @This();

    pub fn initLine(allocator: Allocator, cur: *Parse.Cursor) !Self {
        var self = Self{ .allocator = allocator };
        try self.parseLine(cur);
        return self;
    }

    const ParseError = Allocator.Error || error{
        MissingLeftPacket,
        MissingRightPacket,
        UnexpectedDataLine,
        UnexpectedLineTrailer,
        ExpectedNumber,
        ExpectedListCloseBrace,
    };

    pub fn takePair(self: *Self) !?Pair {
        if (self.left == null and self.right == null) return null;
        const left = self.left orelse return ParseError.MissingLeftPacket;
        const right = self.right orelse return ParseError.MissingRightPacket;
        self.left = null;
        self.right = null;
        return Pair{ .left = left, .right = right };
    }

    pub fn parseLine(self: *Self, cur: *Parse.Cursor) ParseError!void {
        if (std.mem.trim(u8, cur.buf, " \t").len == 0) {
            if (try self.takePair()) |pair|
                try self.pairs.append(self.allocator, pair);
            return;
        }

        if (self.left == null) {
            self.left = try self.parseData(cur);
        } else if (self.right == null) {
            self.right = try self.parseData(cur);
        } else return ParseError.UnexpectedDataLine;

        if (cur.live()) return ParseError.UnexpectedLineTrailer;
    }

    pub fn parseData(self: *Self, cur: *Parse.Cursor) ParseError!Data {
        cur.star(' ');
        return if (cur.have('['))
            self.parseList(cur)
        else
            Data{
                .number = cur.consumeInt(Number, 10) orelse
                    return ParseError.ExpectedNumber,
            };
    }

    pub fn parseList(self: *Self, cur: *Parse.Cursor) ParseError!Data {
        if (cur.have(']')) return Data{ .list = .{} };

        var list = List{};
        errdefer list.deinit(self.allocator);
        while (true) {
            const data = try self.parseData(cur);
            try list.append(self.allocator, data);
            cur.star(' ');
            if (!cur.have(',')) break;
        }
        if (!cur.have(']')) return ParseError.ExpectedListCloseBrace;

        return Data{ .list = list };
    }

    pub fn finish(self: *Self) !World {
        defer self.* = undefined;
        return World{
            .allocator = self.allocator,
            .pairs = self.pairs.toOwnedSlice(self.allocator),
        };
    }
};

const World = struct {
    allocator: Allocator,
    pairs: []Pair,
};

fn run(
    allocator: Allocator,

    // TODO: better "any .reader()-able / any .writer()-able" interfacing
    input: anytype,
    output: anytype,
    config: Config,
) !void {
    var timing = try Timing.start(allocator);
    defer timing.deinit();
    defer timing.printDebugReport();

    var out = output.writer();

    var arena = std.heap.ArenaAllocator.init(allocator);
    defer arena.deinit();

    const world = build: {
        var lines = Parse.lineScanner(input.reader());
        var builder = init: {
            var cur = try lines.next() orelse return error.NoInput;
            break :init Builder.initLine(arena.allocator(), cur) catch |err| return cur.carp(err);
        };
        var lineTime = try timing.timer(.parseLine);
        while (try lines.next()) |cur| {
            builder.parseLine(cur) catch |err| return cur.carp(err);
            try lineTime.lap();
        }
        break :build try builder.finish();
    };
    try timing.markPhase(.parse);

    var orders = try arena.allocator().alloc(math.Order, world.pairs.len);
    var compareTimer = try timing.timer(.compareOne);
    for (world.pairs) |pair, i| {
        orders[i] = Data.order(pair.left, pair.right);
        try compareTimer.lap();
    }
    try timing.markPhase(.compare);

    var oks = try std.DynamicBitSetUnmanaged.initEmpty(arena.allocator(), world.pairs.len);
    for (orders) |ord, i| if (ord != .gt) // .lt or .eq
        oks.set(i);

    const sum = sum_indices: {
        var k: usize = 0;
        var okIt = oks.iterator(.{});
        while (okIt.next()) |i| {
            const n = i + 1;
            k += n;
        }
        break :sum_indices k;
    };

    try timing.markPhase(.count);

    try out.print("# Solution\n", .{});
    if (config.verbose > 0) {
        var okIt = oks.iterator(.{});
        while (okIt.next()) |i| {
            const n = i + 1;
            try out.print("{}. correct\n", .{n});
            if (config.verbose > 1) {
                const pair = world.pairs[i];
                try out.print(
                    \\  - left: {}
                    \\  - right: {}
                    \\
                , .{ pair.left, pair.right });
            }
        }
    }
    try out.print("> {}\n", .{sum});

    try timing.markPhase(.report);

    try timing.finish(.overall);
}

const ArgParser = @import("args.zig").Parser;

const MainAllocator = std.heap.GeneralPurposeAllocator(.{
    // .verbose_log = true,
});

pub fn main() !void {
    var gpa = MainAllocator{};
    defer _ = gpa.deinit();

    var input = std.io.getStdIn();
    var output = std.io.getStdOut();
    var config = Config{};
    var bufferOutput = true;

    {
        var argsArena = std.heap.ArenaAllocator.init(gpa.allocator());
        defer argsArena.deinit();

        var args = try ArgParser.init(argsArena.allocator());
        defer args.deinit();

        // TODO: input filename arg

        while (args.next()) |arg| {
            if (arg.is(.{ "-h", "--help" })) {
                std.debug.print(
                    \\Usage: {s} [-v]
                    \\
                    \\Options:
                    \\
                    \\  -v or
                    \\  --verbose
                    \\    print world state after evaluating each input line
                    \\
                    \\  --raw-output
                    \\    don't buffer stdout writes
                    \\
                , .{args.progName()});
                std.process.exit(0);
            } else if (arg.is(.{ "-v", "--verbose" })) {
                config.verbose += 1;
            } else if (arg.is(.{"--raw-output"})) {
                bufferOutput = false;
            } else return error.InvalidArgument;
        }
    }

    var bufin = std.io.bufferedReader(input.reader());

    if (!bufferOutput)
        return run(gpa.allocator(), &bufin, output, config);

    var bufout = std.io.bufferedWriter(output.writer());
    try run(gpa.allocator(), &bufin, &bufout, config);
    try bufout.flush();
    // TODO: sentinel-buffered output writer to flush lines progressively
    // ... may obviate the desire for raw / non-buffered output else
}
