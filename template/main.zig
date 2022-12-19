const std = @import("std");

const mem = std.mem;

const Allocator = mem.Allocator;

test "example" {
    const example_input =
        \\such data
        \\
    ;

    const test_cases = [_]struct {
        input: []const u8,
        expected: []const u8,
        config: Config,
        skip: bool = false,
    }{
        // Part 1 example
        .{
            .config = .{
                .verbose = 1,
            },
            .input = example_input,
            .expected =
            \\# Solution
            \\> 42
            \\
            ,
        },
    };

    const allocator = std.testing.allocator;

    for (test_cases) |tc, i| {
        if (tc.skip) continue;
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

    solve,

    report,
    overall,
});

const Config = struct {
    verbose: usize = 0,
};

const Builder = struct {
    allocator: Allocator,
    arena: std.heap.ArenaAllocator,
    // TODO state to be built up line-to-line

    const Self = @This();

    pub fn parse(allocator: Allocator, reader: anytype) !World {
        var lines = Parse.lineScanner(reader);
        var builder = init: {
            var cur = try lines.next() orelse return error.NoInput;
            break :init Builder.initLine(allocator, cur) catch |err| return cur.carp(err);
        };
        // var lineTime = try timing.timer(.parseLine);
        while (try lines.next()) |cur| {
            builder.parseLine(cur) catch |err| return cur.carp(err);
            // TODO bring back: try lineTime.lap();
        }
        return builder.finish();
    }

    pub fn initLine(allocator: Allocator, cur: *Parse.Cursor) !Self {
        var self = Self{
            .allocator = allocator,
            .arena = std.heap.ArenaAllocator.init(allocator),
        };
        try self.parseLine(cur);
        return self;
    }

    pub fn parseLine(self: *Self, cur: *Parse.Cursor) !void {
        _ = self;
        if (cur.live()) return error.ParseLineNotImplemented;
    }

    pub fn finish(self: *Self) World {
        // XXX(if may fail): errdefer self.arena.deinit();

        return World{
            .allocator = self.allocator,
            .arena = self.arena,
            // TODO finalized problem data
        };
    }
};

const World = struct {
    allocator: Allocator,
    arena: std.heap.ArenaAllocator,
    // TODO problem representation

    const Self = @This();

    pub fn deinit(self: *Self) void {
        self.arena.deinit();
    }
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

    // FIXME: hookup your config
    _ = config;

    var arena = std.heap.ArenaAllocator.init(allocator);
    defer arena.deinit();

    var world = try Builder.parse(allocator, input.reader());
    defer world.deinit();
    try timing.markPhase(.parse);

    // FIXME: solve...
    try timing.markPhase(.solve);

    try out.print(
        \\# Solution
        \\> {}
        \\
    , .{
        42,
    });
    try timing.markPhase(.report);

    try timing.finish(.overall);
}

const ArgParser = @import("args.zig").Parser;

const MainAllocator = std.heap.GeneralPurposeAllocator(.{
    .enable_memory_limit = true,
    .verbose_log = false,
});

var gpa = MainAllocator{};

var log_level: std.log.Level = .warn;

pub fn log(comptime level: std.log.Level, comptime scope: @TypeOf(.EnumLiteral), comptime format: []const u8, args: anytype) void {
    if (@enumToInt(level) > @enumToInt(log_level)) return;
    const prefix = "[" ++ comptime level.asText() ++ "] (" ++ @tagName(scope) ++ "): ";
    std.debug.getStderrMutex().lock();
    defer std.debug.getStderrMutex().unlock();
    const stderr = std.io.getStdErr().writer();
    nosuspend stderr.print(prefix ++ format ++ "\n", args) catch return;
}

pub fn main() !void {
    gpa.setRequestedMemoryLimit(4 * 1024 * 1024 * 1024);

    var allocator = gpa.allocator();

    var input = std.io.getStdIn();
    var output = std.io.getStdOut();
    var config = Config{};
    var bufferOutput = true;

    {
        var argsArena = std.heap.ArenaAllocator.init(allocator);
        defer argsArena.deinit();

        var args = try ArgParser.init(argsArena.allocator());
        defer args.deinit();

        while (args.next()) |arg| {
            if (!arg.isOption()) {
                var prior = input;
                input = try std.fs.cwd().openFile(arg.have, .{});
                prior.close();
                std.log.info("reading input from {s}", .{arg.have});
            } else if (arg.is(.{ "-h", "--help" })) {
                std.debug.print(
                    \\Usage: {s} [options] [INPUT_FILE]
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
                if (@enumToInt(log_level) < @enumToInt(std.log.Level.debug))
                    log_level = @intToEnum(std.log.Level, @enumToInt(log_level) + 1);
                config.verbose += 1;
            } else if (arg.is(.{"--raw-output"})) {
                bufferOutput = false;
            } else return error.InvalidArgument;
        }
    }

    (do_run: {
        var bufin = std.io.bufferedReader(input.reader());

        if (!bufferOutput)
            break :do_run run(allocator, &bufin, output, config);

        var bufout = std.io.bufferedWriter(output.writer());
        defer bufout.flush() catch {};

        break :do_run run(allocator, &bufin, &bufout, config);
        // TODO: sentinel-buffered output writer to flush lines progressively
        // ... may obviate the desire for raw / non-buffered output else
    }) catch |err| {
        if (err == mem.Allocator.Error.OutOfMemory) {
            const have_leaks = gpa.detectLeaks();
            std.log.err("{} detectLeaks() -> {}", .{ err, have_leaks });
        }
        return err;
    };
    if (gpa.detectLeaks()) return error.LeakedMemory;
}
