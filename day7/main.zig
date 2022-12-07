const std = @import("std");
const Allocator = std.mem.Allocator;

test "example" {
    const example =
        \\$ cd /
        \\$ ls
        \\dir a
        \\14848514 b.txt
        \\8504156 c.dat
        \\dir d
        \\$ cd a
        \\$ ls
        \\dir e
        \\29116 f
        \\2557 g
        \\62596 h.lst
        \\$ cd e
        \\$ ls
        \\584 i
        \\$ cd ..
        \\$ cd ..
        \\$ cd d
        \\$ ls
        \\4060174 j
        \\8033020 d.log
        \\5626152 d.ext
        \\7214296 k
        \\
    ;

    const expected =
        \\+ 94853 /a
        \\+ 584 /a/e
        \\> 95437
        \\
    ;

    const allocator = std.testing.allocator;

    var input = std.io.fixedBufferStream(example);
    var output = std.ArrayList(u8).init(allocator);
    defer output.deinit();

    run(allocator, &input, &output) catch |err| {
        std.debug.print("```pre-error output:\n{s}\n```\n", .{output.items});
        return err;
    };
    try std.testing.expectEqualStrings(expected, output.items);
}

const Entry = union(enum) {
    file: File,
    dir: Dir,

    const File = struct {
        name: []u8,
        size: usize,
    };

    const Dir = struct {
        name: []u8,
        first: ?*List,

        // fn getent(self: Dir, name: []u8) ?Entry {}
        // fn mkdir(self: *Dir, name: []u8, allocator: Allocator) !Dir {}
        // fn touch(self: *Dir, name: []u8, allocator: Allocator) !File {}

    };

    const List = struct {
        ent: Entry,
        next: ?*List,
    };
};

const Parse = @import("./parse.zig");

fn run(
    under_allocator: Allocator,

    // TODO: better "any .reader()-able / any .writer()-able" interfacing
    input: anytype,
    output: anytype,
) !void {
    var arena = std.heap.ArenaAllocator.init(under_allocator);
    defer arena.deinit();

    // FIXME: maybe use this
    // const allocator = arena.allocator();

    var lines = Parse.lineScanner(input.reader());
    var out = output.writer();

    while (try lines.next()) |*cur| {
        if (cur.have('$')) {
            cur.expectStar(' ');
            try out.print("* TODO parse command `{s}`\n", .{cur.rem()});
        } else try out.print("! unexpected line `{s}`\n", .{cur.rem()});
    }

    // FIXME: very answer
    try out.print("> {}\n", .{42});
}

pub fn main() !void {
    const allocator = std.heap.page_allocator;

    var input = std.io.getStdIn();
    var output = std.io.getStdOut();

    var bufin = std.io.bufferedReader(input.reader());
    var bufout = std.io.bufferedWriter(output.writer());

    try run(allocator, &bufin, &bufout);
    try bufout.flush();

    // TODO: argument parsing to steer input selection

    // TODO: sentinel-buffered output writer to flush lines progressively

    // TODO: input, output, and run-time metrics
}
