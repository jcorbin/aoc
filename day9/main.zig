const std = @import("std");
const Allocator = std.mem.Allocator;

test "example" {
    const test_cases = [_]struct {
        input: []const u8,
        expected: []const u8,
        config: Config,
    }{
        // Part 1 example
        .{
            .config = .{
                .verbose = true,
                .bounds = .{
                    .from = .{ .x = 0, .y = -4 },
                    .to = .{ .x = 6, .y = 1 },
                },
            },
            .input = 
            \\R 4
            \\U 4
            \\L 3
            \\D 1
            \\R 4
            \\D 1
            \\L 5
            \\R 2
            \\
            ,
            .expected = 
            \\# Eval 1. R 4
            \\    ......
            \\    ......
            \\    ......
            \\    ......
            \\    s..TH.
            \\
            \\# Eval 2. U 4
            \\    ....H.
            \\    ....T.
            \\    ......
            \\    ......
            \\    s.....
            \\
            \\# Eval 3. L 3
            \\    .HT...
            \\    ......
            \\    ......
            \\    ......
            \\    s.....
            \\
            \\# Eval 4. D 1
            \\    ..T...
            \\    .H....
            \\    ......
            \\    ......
            \\    s.....
            \\
            \\# Eval 5. R 4
            \\    ......
            \\    ....TH
            \\    ......
            \\    ......
            \\    s.....
            \\
            \\# Eval 6. D 1
            \\    ......
            \\    ....T.
            \\    .....H
            \\    ......
            \\    s.....
            \\
            \\# Eval 7. L 5
            \\    ......
            \\    ......
            \\    HT....
            \\    ......
            \\    s.....
            \\
            \\# Eval 8. R 2
            \\    ......
            \\    ......
            \\    .TH...
            \\    ......
            \\    s.....
            \\
            \\# Tail Visited
            \\    ..##..
            \\    ...##.
            \\    .####.
            \\    ....#.
            \\    s###..
            \\
            \\> 13
            \\
            ,
        },

        // Part 2 large example
        .{
            .config = .{
                .verbose = true,
                .tailKnots = 9,
                .bounds = .{
                    .from = .{ .x = -11, .y = -15 },
                    .to = .{ .x = 15, .y = 6 },
                },
            },
            .input = 
            \\R 5
            \\U 8
            \\L 8
            \\D 3
            \\R 17
            \\D 10
            \\L 25
            \\U 20
            \\
            ,
            .expected = 
            \\# Eval 1. R 5
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ...........54321H.........
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\
            \\# Eval 2. U 8
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ................H.........
            \\    ................1.........
            \\    ................2.........
            \\    ................3.........
            \\    ...............54.........
            \\    ..............6...........
            \\    .............7............
            \\    ............8.............
            \\    ...........9..............
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\
            \\# Eval 3. L 8
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ........H1234.............
            \\    ............5.............
            \\    ............6.............
            \\    ............7.............
            \\    ............8.............
            \\    ............9.............
            \\    ..........................
            \\    ..........................
            \\    ...........s..............
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\
            \\# Eval 4. D 3
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    .........2345.............
            \\    ........1...6.............
            \\    ........H...7.............
            \\    ............8.............
            \\    ............9.............
            \\    ..........................
            \\    ..........................
            \\    ...........s..............
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\
            \\# Eval 5. R 17
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ................987654321H
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ...........s..............
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\
            \\# Eval 6. D 10
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ...........s.........98765
            \\    .........................4
            \\    .........................3
            \\    .........................2
            \\    .........................1
            \\    .........................H
            \\
            \\# Eval 7. L 25
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ...........s..............
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    H123456789................
            \\
            \\# Eval 8. U 20
            \\    H.........................
            \\    1.........................
            \\    2.........................
            \\    3.........................
            \\    4.........................
            \\    5.........................
            \\    6.........................
            \\    7.........................
            \\    8.........................
            \\    9.........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ...........s..............
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\
            \\# Tail Visited
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    ..........................
            \\    #.........................
            \\    #.............###.........
            \\    #............#...#........
            \\    .#..........#.....#.......
            \\    ..#..........#.....#......
            \\    ...#........#.......#.....
            \\    ....#......s.........#....
            \\    .....#..............#.....
            \\    ......#............#......
            \\    .......#..........#.......
            \\    ........#........#........
            \\    .........########.........
            \\
            \\> 36
            \\
            ,
        },
    };

    const allocator = std.testing.allocator;

    for (test_cases) |tc| {
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

const Parse = @import("./parse.zig");
const Timing = @import("./perf.zig").Timing;

// TODO really should sort out how to Vector
const Point = struct {
    x: i32,
    y: i32,

    const Self = @This();

    pub fn of(x: i32, y: i32) Point {
        return .{ .x = x, .y = y };
    }

    pub fn add(a: Self, b: Self) Self {
        return .{ .x = a.x + b.x, .y = a.y + b.y };
    }

    pub fn sub(a: Self, b: Self) Self {
        return .{ .x = a.x - b.x, .y = a.y - b.y };
    }

    pub fn absDiff(a: Self, b: Self) struct { x: u32, y: u32 } {
        return .{
            .x = @intCast(u32, if (a.x > b.x) a.x - b.x else b.x - a.x),
            .y = @intCast(u32, if (a.y > b.y) a.y - b.y else b.y - a.y),
        };
    }

    pub fn indexInto(self: Self, offset: usize, stride: usize) usize {
        return offset + stride * @intCast(usize, self.y) + @intCast(usize, self.x);
    }
};

const Rect = struct {
    from: Point,
    to: Point,

    const Self = @This();

    pub fn width(self: Self) u32 {
        const x2 = self.to.x;
        const x1 = self.from.x;
        return if (x2 > x1) @intCast(u32, x2 - x1) else 0;
    }

    pub fn height(self: Self) u32 {
        const y2 = self.to.y;
        const y1 = self.from.y;
        return if (y2 > y1) @intCast(u32, y2 - y1) else 0;
    }

    pub fn contains(self: Self, p: Point) bool {
        const x2 = self.to.x;
        const x1 = self.from.x;
        const y2 = self.to.y;
        const y1 = self.from.y;
        return x1 <= p.x and p.x < x2 and y1 <= p.y and p.y < y2;
    }

    pub fn expandTo(self: *Self, p: Point) void {
        if (self.to.x > self.from.x and self.to.y > self.from.y) {
            if (p.x < self.from.x) self.from.x = p.x;
            if (p.y < self.from.y) self.from.y = p.y;
            if (p.x >= self.to.x) self.to.x = p.x + 1;
            if (p.y >= self.to.y) self.to.y = p.y + 1;
        } else {
            self.from.x = p.x;
            self.from.y = p.y;
            self.to.x = p.x + 1;
            self.to.y = p.y + 1;
        }
    }
};

fn Layer(comptime Datum: type) type {
    return struct {
        const Index = std.AutoHashMap(Point, usize);

        // TODO: evaluate switching to MultiArrayList
        const Data = std.ArrayList(struct {
            loc: Point,
            tag: Datum,
        });

        bounds: Rect = .{
            .from = .{ .x = 0, .y = 0 },
            .to = .{ .x = 0, .y = 0 },
        },
        index: Index,
        data: Data,

        const RegionIterator = struct {
            const It = @This();

            items: Self.Data.Slice,
            bounds: Rect,
            at: Point = .{ .x = 0, .y = 0 },
            i: usize = 0,

            pub fn next(it: *It) ?usize {
                var i = it.i;
                while (i < it.items.len) : (i += 1) {
                    const loc = it.items[i].loc;
                    if (it.bounds.contains(loc)) {
                        it.at = loc;
                        it.i = i + 1;
                        return i;
                    }
                }
                return null;
            }
        };

        const Self = @This();

        pub fn init(allocator: Allocator) Self {
            return .{
                .index = Index.init(allocator),
                .data = Data.init(allocator),
            };
        }

        pub fn deinit(self: *Self) void {
            self.index.deinit();
            self.data.deinit();
        }

        pub fn move(self: *Self, from: Point, to: Point) !void {
            const kv = self.index.fetchRemove(from) orelse return;
            const i = kv.value;
            try self.index.put(to, i);
            self.data.items[i].loc = to;
            self.bounds.expandTo(to); // NOTE: bounds doesn't shrink ex from
        }

        pub fn set(self: *Self, loc: Point, tag: Datum) !void {
            if (self.index.get(loc)) |i| {
                self.data.items[i].tag = tag;
            } else {
                const i = self.data.items.len;
                try self.data.append(.{
                    .loc = loc,
                    .tag = tag,
                });
                try self.index.put(loc, i);

                self.bounds.expandTo(loc);
            }
        }

        pub fn get(self: *Self, loc: Point) ?Datum {
            return if (self.index.get(loc)) |i| self.data.items[i].tag else null;
        }

        pub fn withinRegion(self: *Self, bounds: Rect) RegionIterator {
            return .{
                .items = self.data.items,
                .bounds = bounds,
            };
        }
    };
}

const Tag = union(enum) {
    mark: enum {
        empty,
        start,
        visited,
        head,
        tail,

        pub fn glyph(self: @This()) u8 {
            return switch (self) {
                .empty => '.',
                .start => 's',
                .visited => '#',
                .head => 'H',
                .tail => 'T',
            };
        }
    },
    knot: u4,

    pub fn glyph(self: @This()) u8 {
        return switch (self) {
            .mark => |m| m.glyph(),
            .knot => |k| if (k < 10)
                '0' + @intCast(u8, k)
            else
                'A' + @intCast(u8, k),
        };
    }
};

const World = struct {
    const Rope = Layer(Tag).Data;

    arena: std.heap.ArenaAllocator,
    tmpString: std.ArrayList(u8),

    background: Layer(Tag),
    rope: Rope,

    const Self = @This();

    pub fn init(allocator: Allocator) Self {
        var self = Self{
            .arena = std.heap.ArenaAllocator.init(allocator),
            .background = Layer(Tag).init(allocator),
            .rope = Rope.init(allocator),
            .tmpString = std.ArrayList(u8).init(allocator),
        };
        return self;
    }

    pub fn deinit(self: *Self) void {
        self.tmpString.deinit();
        self.background.deinit();
        self.rope.deinit();
        self.arena.deinit();
    }

    pub fn startAt(self: *Self, at: Point, tailKnots: u4) !void {
        try self.background.set(at, .{ .mark = .start });
        try self.rope.ensureTotalCapacity(1 + tailKnots);

        try self.rope.append(.{
            .loc = at,
            .tag = .{ .mark = .head },
        });
        switch (tailKnots) {
            0 => {},
            1 => try self.rope.append(.{
                .loc = at,
                .tag = .{ .mark = .tail },
            }),
            else => {
                const n = @intCast(u8, tailKnots);
                var i: u8 = 1;
                while (i <= n) : (i += 1)
                    try self.rope.append(.{
                        .loc = at,
                        .tag = .{ .knot = @intCast(u4, i) },
                    });
            },
        }
    }

    pub fn evalMove(self: *Self, cur: *Parse.Cursor) !void {
        const moveCode = cur.consume() orelse return error.MissingMoveCode;
        const moveDir = switch (moveCode) {
            'R' => Point.of(1, 0),
            'L' => Point.of(-1, 0),
            'U' => Point.of(0, -1),
            'D' => Point.of(0, 1),
            else => return error.InvalidMoveCode,
        };

        try cur.expect(' ', error.MissingMoveSpace);

        const moveCount = try cur.expectInt(u8, error.MissingMoveCount);
        if (moveCount < 1) return error.InvalidMoveCount;

        var i: u8 = 0;
        while (i < moveCount) : (i += 1)
            try self.moveHeadBy(moveDir);
    }

    pub fn moveHeadBy(self: *Self, by: Point) !void {
        const head = &self.rope.items[0];
        var to = head.loc.add(by);
        head.loc = to;

        for (self.rope.items[1..]) |*item| {
            var loc = item.loc;

            // loc follows to, from head on down
            //
            // T . H   . T H
            // . . . > . . .
            // . . .   . . .
            //
            // . . H   . . H
            // . . . > . . T
            // . . T   . . .
            //
            // . . H   . T H
            // T . . > . . .
            // . . .   . . .
            //
            // . . H   . . H
            // . . . > . . T
            // . T .   . . .

            const d = to.sub(loc);

            const distX = std.math.absInt(d.x) catch 9; // 9 chosen to ensure `else` below
            const distY = std.math.absInt(d.y) catch 9; // ... and because dilbert of course
            if (distX > 1 or distY > 1) {
                if (d.x > 0) {
                    loc.x += 1;
                } else if (d.x < 0) {
                    loc.x -= 1;
                }
                if (d.y > 0) {
                    loc.y += 1;
                } else if (d.y < 0) {
                    loc.y -= 1;
                }
            }

            item.loc = loc;
            to = loc;
        }

        // mark background visited at the rope tail tip
        switch (self.background.get(to) orelse Tag{ .mark = .empty }) {
            .mark => |m| switch (m) {
                .empty => try self.background.set(to, .{ .mark = .visited }),
                else => {},
            },
            else => {},
        }
    }

    pub fn toString(self: *Self, config: struct {
        prefix: []const u8 = "",
        showStart: bool = false,
        showVisited: bool = false,
        showRope: bool = false,
    }) ![]u8 {
        const bounds = self.background.bounds;
        const offset = config.prefix.len; // no extra sapce at start of line
        const stride = offset + bounds.width() + 1; // space for newline terminator

        const memSize = bounds.height() * stride - 1; // less final newline
        try self.tmpString.resize(memSize);
        var buf = self.tmpString.items;

        // fill empty
        std.mem.set(u8, buf, (Tag{ .mark = .empty }).glyph());

        { // prefix lines
            var i: usize = 0;
            while (i < buf.len) : (i += stride)
                std.mem.copy(u8, buf[i..], config.prefix);
        }

        { // terminate lines
            var i = stride - 1;
            while (i < buf.len) : (i += stride)
                buf[i] = '\n';
        }

        if (config.showStart or config.showVisited) {
            var it = self.background.withinRegion(bounds);
            while (it.next()) |i| {
                const j = it.at.sub(bounds.from).indexInto(offset, stride);
                const tag = self.background.data.items[i].tag;
                switch (tag) {
                    .mark => |m| switch (m) {
                        .start => {
                            if (config.showStart) buf[j] = tag.glyph();
                        },
                        .visited => {
                            if (config.showVisited) buf[j] = tag.glyph();
                        },
                        else => {},
                    },
                    else => {},
                }
            }
        }

        if (config.showRope) {
            var i: usize = self.rope.items.len;
            while (i > 0) {
                i -= 1;
                const item = &self.rope.items[i];
                if (!self.background.bounds.contains(item.loc)) continue;
                const j = item.loc.sub(bounds.from).indexInto(offset, stride);
                buf[j] = item.tag.glyph();
            }
        }

        return self.tmpString.items;
    }
};

const Config = struct {
    bounds: Rect = .{
        .from = .{ .x = 0, .y = 0 },
        .to = .{ .x = 0, .y = 0 },
    },
    verbose: bool = false,
    tailKnots: u4 = 1,
};

fn run(
    allocator: Allocator,

    // TODO: better "any .reader()-able / any .writer()-able" interfacing
    input: anytype,
    output: anytype,
    config: Config,
) !void {
    var timing = try Timing(enum {
        eval,
        evalLine,
        evalShow,
        reportVisited,
        overall,
    }).start(allocator);
    defer timing.deinit();
    defer timing.printDebugReport();

    var world = World.init(allocator);
    defer world.deinit();
    world.background.bounds = config.bounds;

    try world.startAt(.{ .x = 0, .y = 0 }, config.tailKnots);

    var lines = Parse.lineScanner(input.reader());
    var out = output.writer();

    // evaluate moves
    while (try lines.next()) |*cur| {
        var lineTime = try std.time.Timer.start();
        world.evalMove(cur) catch |err| {
            std.debug.print("failed to eval line #{}: `{s}`\n", .{ cur.count, cur.buf });
            return err;
        };
        try timing.collect(.evalLine, lineTime.lap());

        if (config.verbose) {
            try out.print(
                \\# Eval {}. {s}
                \\{s}
                \\
                \\
            , .{
                cur.count,
                cur.buf,
                try world.toString(.{
                    .prefix = "    ",
                    .showRope = true,
                    .showStart = true,
                }),
            });
            try timing.collect(.evalShow, lineTime.lap());
        }
    }
    try timing.markPhase(.eval);

    // How many positions does the tail of the rope visit at least once?
    {
        var count: usize = 0;
        for (world.background.data.items) |item| {
            switch (item.tag) {
                .mark => |m| switch (m) {
                    .start, .visited => count += 1,
                    else => {},
                },
                else => {},
            }
        }

        try out.print(
            \\# Tail Visited
            \\{s}
            \\
            \\> {}
            \\
        , .{
            try world.toString(.{
                .prefix = "    ",
                .showStart = true,
                .showVisited = true,
            }),
            count,
        });
    }
    try timing.markPhase(.reportVisited);

    try timing.finish(.overall);
}

const ArgParser = @import("./args.zig").Parser;

pub fn main() !void {
    const allocator = std.heap.page_allocator;

    var input = std.io.getStdIn();
    var output = std.io.getStdOut();
    var config = Config{};

    {
        var args = try ArgParser.init(allocator);
        defer args.deinit();

        // TODO: input filename arg

        while (try args.next()) |arg| {
            if (arg.is(.{ "-h", "--help" })) {
                std.debug.print(
                    \\Usage: {s} [-v] [-k NUMBER]
                    \\
                    \\Options:
                    \\  -v or
                    \\  --verbose
                    \\    print world state after evaluating each input line
                    \\
                    \\  -k NUMBER or
                    \\  --tail-knots NUMBER
                    \\    how many knots follow the main rope head knot
                    \\    defaults to 1
                    \\
                , .{args.progName()});
                std.process.exit(0);
            } else if (arg.is(.{ "-v", "--verbose" })) {
                config.verbose = true;
            } else if (arg.is(.{ "-k", "--tail-knots" })) {
                var valueArg = (try args.next()) orelse return error.MissingKnotsValue;
                config.tailKnots = try valueArg.parseInt(u4, 10);
            } else return error.InvalidArgument;
        }
    }

    var bufin = std.io.bufferedReader(input.reader());
    var bufout = std.io.bufferedWriter(output.writer());

    try run(allocator, &bufin, &bufout, config);
    try bufout.flush();

    // TODO: sentinel-buffered output writer to flush lines progressively
}
