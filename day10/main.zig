const std = @import("std");
const Allocator = std.mem.Allocator;

test "example" {
    const large_sample =
        \\addx 15
        \\addx -11
        \\addx 6
        \\addx -3
        \\addx 5
        \\addx -1
        \\addx -8
        \\addx 13
        \\addx 4
        \\noop
        \\addx -1
        \\addx 5
        \\addx -1
        \\addx 5
        \\addx -1
        \\addx 5
        \\addx -1
        \\addx 5
        \\addx -1
        \\addx -35
        \\addx 1
        \\addx 24
        \\addx -19
        \\addx 1
        \\addx 16
        \\addx -11
        \\noop
        \\noop
        \\addx 21
        \\addx -15
        \\noop
        \\noop
        \\addx -3
        \\addx 9
        \\addx 1
        \\addx -3
        \\addx 8
        \\addx 1
        \\addx 5
        \\noop
        \\noop
        \\noop
        \\noop
        \\noop
        \\addx -36
        \\noop
        \\addx 1
        \\addx 7
        \\noop
        \\noop
        \\noop
        \\addx 2
        \\addx 6
        \\noop
        \\noop
        \\noop
        \\noop
        \\noop
        \\addx 1
        \\noop
        \\noop
        \\addx 7
        \\addx 1
        \\noop
        \\addx -13
        \\addx 13
        \\addx 7
        \\noop
        \\addx 1
        \\addx -33
        \\noop
        \\noop
        \\noop
        \\addx 2
        \\noop
        \\noop
        \\noop
        \\addx 8
        \\noop
        \\addx -1
        \\addx 2
        \\addx 1
        \\noop
        \\addx 17
        \\addx -9
        \\addx 1
        \\addx 1
        \\addx -3
        \\addx 11
        \\noop
        \\noop
        \\addx 1
        \\noop
        \\addx 1
        \\noop
        \\noop
        \\addx -13
        \\addx -19
        \\addx 1
        \\addx 3
        \\addx 26
        \\addx -30
        \\addx 12
        \\addx -1
        \\addx 3
        \\addx 1
        \\noop
        \\noop
        \\noop
        \\addx -9
        \\addx 18
        \\addx 1
        \\addx 2
        \\noop
        \\noop
        \\addx 9
        \\noop
        \\noop
        \\noop
        \\addx -1
        \\addx 2
        \\addx -37
        \\addx 1
        \\addx 3
        \\noop
        \\addx 15
        \\addx -21
        \\addx 22
        \\addx -6
        \\addx 1
        \\noop
        \\addx 2
        \\addx 1
        \\noop
        \\addx -10
        \\noop
        \\noop
        \\addx 20
        \\addx 1
        \\addx 2
        \\addx 2
        \\addx -6
        \\addx -11
        \\noop
        \\noop
        \\noop
        \\
    ;
    const test_cases = [_]struct {
        input: []const u8,
        expected: []const u8,
        config: Config,
    }{
        // Part 1 small example
        .{
            .config = .{
                .verbose = true,
            },
            .input = 
            \\noop
            \\addx 3
            \\addx -5
            \\
            ,
            .expected = 
            \\# Eval 1. `noop`
            \\    cycle: 1 x: 1
            \\# Eval 2. `addx 3`
            \\    cycle: 2 x: 1
            \\    cycle: 3 x: 1
            \\# Eval 3. `addx -5`
            \\    cycle: 4 x: 4
            \\    cycle: 5 x: 4
            \\# Halt
            \\    cycle: 6 x: -1
            \\
            ,
        },

        // Part 1 large example
        .{
            .config = .{
                .signalFrom = 20,
                .signalEvery = 40,
            },
            .input = large_sample,
            // The interesting signal strengths can be determined as follows:
            //
            // - During the 20th cycle, register X has the value 21, so the signal strength is 20 * 21 = *`420`*.
            //   (The 20th cycle occurs in the middle of the second `addx -1`, so the value of
            //   register X is the starting value, 1, plus all of the other `addx` values up to
            //   that point: 1 + 15 - 11 + 6 - 3 + 5 - 1 - 8 + 13 + 4 = 21.)
            // - During the 60th cycle, register X has the value 19, so the signal strength is 60 * 19 = *`1140`*.
            // - During the 100th cycle, register X has the value 18, so the signal strength is 100 * 18 = *`1800`*.
            // - During the 140th cycle, register X has the value 21, so the signal strength is 140 * 21 = *`2940`*.
            // - During the 180th cycle, register X has the value 16, so the signal strength is 180 * 16 = *`2880`*.
            // - During the 220th cycle, register X has the value 18, so the signal strength is 220 * 18 = *`3960`*.
            //
            // The sum of these signal strengths is *`13140`*.
            .expected = 
            \\# cycle 20
            \\   x: 21
            \\   signal: 420
            \\> 420
            \\
            \\# cycle 60
            \\   x: 19
            \\   signal: 1140
            \\> 1560
            \\
            \\# cycle 100
            \\   x: 18
            \\   signal: 1800
            \\> 3360
            \\
            \\# cycle 140
            \\   x: 21
            \\   signal: 2940
            \\> 6300
            \\
            \\# cycle 180
            \\   x: 16
            \\   signal: 2880
            \\> 9180
            \\
            \\# cycle 220
            \\   x: 18
            \\   signal: 3960
            \\> 13140
            \\
            ,
        },

        // Part 2 large example
        .{
            .config = .{
                .crt = .{ .on = .{ .width = 40, .height = 6 } },
            },
            .input = large_sample,
            .expected = 
            \\# Frame 1
            \\    ##..##..##..##..##..##..##..##..##..##..
            \\    ###...###...###...###...###...###...###.
            \\    ####....####....####....####....####....
            \\    #####.....#####.....#####.....#####.....
            \\    ######......######......######......####
            \\    #######.......#######.......#######.....
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

const CPU = struct {
    cycle: usize = 1,
    x: i64 = 1,

    const Op = union(enum) {
        noop: void,
        addx: i64,

        pub fn parse(buf: []const u8) !Op {
            return if (std.mem.eql(u8, buf, "noop"))
                Op{ .noop = {} }
            else if (std.mem.startsWith(u8, buf, "addx "))
                Op{ .addx = try std.fmt.parseInt(i64, buf[5..], 10) }
            else
                error.UnrecognizedOp;
        }
    };

    const Self = @This();

    pub fn exec(self: Self, op: Op) Self {
        switch (op) {
            .noop => return .{ .cycle = self.cycle + 1, .x = self.x },
            .addx => |n| return .{ .cycle = self.cycle + 2, .x = self.x + n },
        }
    }
};

const CRT = struct {
    allocator: Allocator,

    width: usize,
    height: usize,

    frame: []u8,
    frameCount: usize = 0,

    lineOffset: usize,
    lineStride: usize,

    rayX: usize = 0,
    rayY: usize = 0,

    dark: u8 = '.',
    sprite: []const u8 = "###",
    spriteX: i64 = 1,

    const Self = @This();

    pub fn init(
        allocator: Allocator,
        width: usize,
        height: usize,
        lineStart: []const u8,
        lineEnd: []const u8,
    ) !Self {
        const stride = lineStart.len + width + lineEnd.len;
        var self = Self{
            .allocator = allocator,
            .width = width,
            .height = height,
            .frame = try allocator.alloc(u8, stride * height),
            .lineOffset = lineStart.len,
            .lineStride = stride,
        };

        std.mem.set(u8, self.frame, '_');

        var i: usize = 0;
        while (i < self.frame.len) : (i += self.lineStride)
            std.mem.copy(u8, self.frame[i..], lineStart);

        i = self.lineOffset + self.width;
        while (i < self.frame.len) : (i += self.lineStride)
            std.mem.copy(u8, self.frame[i..], lineEnd);

        return self;
    }

    pub fn deinit(self: *Self) void {
        self.allocator.free(self.frame);
    }

    pub fn spriteAt(self: *Self) i64 {
        return self.spriteX - @intCast(i64, @divTrunc(self.sprite.len, 2));
    }

    pub fn copySpriteInto(self: *Self, buf: []u8) void {
        var at = self.spriteAt();
        const w = std.math.min(
            self.sprite.len,
            @intCast(i64, buf.len) - at,
        );

        var o: usize = 0;
        if (at < 0) {
            o += @intCast(usize, -at);
            at = 0;
            if (o >= w) return;
        }

        if (w > 0)
            std.mem.copy(u8, buf[@intCast(usize, at)..], self.sprite[o..w]);
    }

    pub fn pixel(self: *Self) u8 {
        const i = @intCast(i64, self.rayX) - self.spriteAt();
        return if (0 <= i and i < @intCast(i64, self.sprite.len))
            self.sprite[@intCast(usize, i)]
        else
            self.dark;
    }

    const Frame = struct {
        count: usize,
        buf: []const u8,
    };

    pub fn advance(self: *Self) ?Frame {
        self.frame[
            self.lineOffset + self.rayY * self.lineStride + self.rayX
        ] = self.pixel();

        self.rayX += 1;
        if (self.rayX < self.width) return null;
        self.rayX -= self.width;

        self.rayY += 1;
        if (self.rayY < self.height) return null;
        self.rayY -= self.height;

        self.frameCount += 1;
        return Frame{
            .count = self.frameCount,
            .buf = self.trimFrame(),
        };
    }

    pub fn trimFrame(self: *Self) []const u8 {
        return std.mem.trimRight(u8, self.frame, "\n");
    }
};

fn Signal(comptime Writer: type) type {
    return struct {
        writer: Writer,
        from: usize,
        every: usize,
        total: i64 = 0,

        const Self = @This();

        pub fn collect(self: *Self, cpu: CPU) void {
            if (cpu.cycle < self.from) return;
            if ((cpu.cycle - self.from) % self.every == 0) {
                const signal = @intCast(i64, cpu.cycle) * cpu.x;
                self.total += signal;

                if (cpu.cycle > self.from)
                    self.writer.writeByte('\n') catch return;
                self.writer.print(
                    \\# cycle {}
                    \\   x: {}
                    \\   signal: {}
                    \\> {}
                    \\
                , .{
                    cpu.cycle,
                    cpu.x,
                    signal,
                    self.total,
                }) catch return;
            }
        }
    };
}

const Parse = @import("./parse.zig");
const Timing = @import("./perf.zig").Timing;

const Config = struct {
    verbose: bool = false,
    debug: bool = false,
    signalFrom: usize = 0,
    signalEvery: usize = 0,
    crt: union(enum) {
        off: void,
        on: struct {
            width: usize,
            height: usize,
        },
    } = .{ .off = {} },
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
        evalLineVerbose,
        overall,
    }).start(allocator);
    defer timing.deinit();
    defer timing.printDebugReport();

    var lines = Parse.lineScanner(input.reader());
    var out = output.writer();

    var cpu = CPU{};

    var signal = if (config.signalEvery > 0) Signal(@TypeOf(out)){
        .writer = out,
        .from = config.signalFrom,
        .every = config.signalEvery,
    } else null;

    var crt = switch (config.crt) {
        .off => null,
        .on => |cfg| try CRT.init(allocator, cfg.width, cfg.height, "    ", "\n"),
    };
    defer if (crt) |*c| c.deinit();

    // evaluate input

    if (config.debug) {
        if (crt) |*c| {
            var tmp = [_]u8{'.'} ** 40;
            c.copySpriteInto(tmp[0..]);
            std.debug.print(
                \\Sprite position: {s}
                \\
                \\
            , .{tmp});
        }
    }

    while (try lines.next()) |*cur| {
        var lineTime = try std.time.Timer.start();
        const op = try CPU.Op.parse(cur.buf);

        if (config.verbose) try out.print(
            \\# Eval {}. `{s}`
            \\
        , .{ cur.count, cur.buf });

        const next = cpu.exec(op);

        if (config.debug)
            std.debug.print(
                \\Start cycle  {}: begin executing {} (X: {})
                \\
            , .{ cpu.cycle, op, cpu.x });

        while (cpu.cycle < next.cycle) : (cpu.cycle += 1) {
            if (config.verbose) try out.print(
                \\    cycle: {} x: {}
                \\
            , .{ cpu.cycle, cpu.x });

            if (signal) |*sig| sig.collect(cpu);
            if (crt) |*c| {
                if (config.debug)
                    std.debug.print(
                        \\During cycle {}: CRT at {}, {}
                        \\
                    , .{ cpu.cycle, c.rayX, c.rayY });

                if (c.advance()) |frame| try out.print(
                    \\# Frame {}
                    \\{s}
                    \\
                , .{ frame.count, frame.buf });
            }
        }
        cpu = next;

        if (config.debug)
            std.debug.print(
                \\End of cycle {}: finish executing {} (X: {})
                \\
            , .{ cpu.cycle - 1, op, cpu.x });

        if (crt) |*c| {
            c.spriteX = cpu.x;

            if (config.debug) {
                var tmp = [_]u8{'.'} ** 40;
                c.copySpriteInto(tmp[0..]);
                std.debug.print(
                    \\Sprite position: {s}
                    \\
                    \\
                , .{tmp});
            }
        }

        try timing.collect(.evalLine, lineTime.lap());
    }

    if (config.verbose) try out.print(
        \\# Halt
        \\    cycle: {} x: {}
        \\
    , .{ cpu.cycle, cpu.x });

    if (signal) |*sig| sig.collect(cpu);

    try timing.markPhase(.eval);

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
                    \\Usage: {s} [-v]
                    \\
                    \\Options:
                    \\  -v or
                    \\  --verbose
                    \\    print world state after evaluating each input line
                    \\
                    \\  -s FROM EVERY or
                    \\  --signal FROM EVERY
                    \\    collect signal after FROM cycles and then EVERY cycles thereafter
                    \\
                    \\  -c WIDTH HEIGHT or
                    \\  --crt WIDTH HEIGHT 
                    \\    collect signal after FROM cycles and then EVERY cycles thereafter
                    \\
                , .{args.progName()});
                std.process.exit(0);
            } else if (arg.is(.{ "-s", "--signal" })) {
                var fromArg = (try args.next()) orelse return error.MissingFromValue;
                var everyArg = (try args.next()) orelse return error.MissingEveryValue;
                config.signalFrom = try fromArg.parseInt(usize, 10);
                config.signalEvery = try everyArg.parseInt(usize, 10);
            } else if (arg.is(.{ "-c", "--crt" })) {
                var widthArg = (try args.next()) orelse return error.MissingWidthValue;
                var heightArg = (try args.next()) orelse return error.MissingHeightValue;
                config.crt = .{ .on = .{
                    .width = try widthArg.parseInt(usize, 10),
                    .height = try heightArg.parseInt(usize, 10),
                } };
            } else if (arg.is(.{ "-v", "--verbose" })) {
                if (!config.verbose) {
                    config.verbose = true;
                } else {
                    config.debug = true;
                }
            } else return error.InvalidArgument;
        }
    }

    var bufin = std.io.bufferedReader(input.reader());
    var bufout = std.io.bufferedWriter(output.writer());

    try run(allocator, &bufin, &bufout, config);
    try bufout.flush();

    // TODO: sentinel-buffered output writer to flush lines progressively
}
