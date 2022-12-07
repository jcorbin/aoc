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
    const Self = @This();

    file: File,
    dir: Dir,

    pub fn name(self: *Self) []const u8 {
        return switch (self.*) {
            .dir => |d| d.name,
            .file => |f| f.name,
        };
    }
};

const File = struct {
    name: []const u8,
    size: usize,
};

const Dir = struct {
    const Self = @This();

    name: []const u8,
    parent: ?*Dir = null,
    first: ?*ListNode = null,

    const ListNode = struct {
        ent: Entry,
        next: ?*ListNode = null,
    };

    pub fn getent(self: *Self, name: []const u8) ?*Entry {
        var next = self.first;
        while (next) |list| {
            if (std.mem.eql(u8, list.ent.name(), name)) return &list.ent;
            next = list.next;
        }
        return null;
    }

    pub fn addent(self: *Dir, item: *ListNode) !void {
        const itemName = item.ent.name();

        var prior: ?*ListNode = null;
        var next = self.first;
        while (next) |list| {
            const entName = list.ent.name();
            if (!std.mem.lessThan(u8, entName, itemName)) {
                if (std.mem.eql(u8, entName, itemName)) return error.DuplicateEntryName;
                break;
            }
            prior = next;
            next = list.next;
        }

        item.next = next;
        if (prior) |prev| prev.next = item else self.first = item;
    }
};

const Device = struct {
    const Self = @This();

    arena: std.heap.ArenaAllocator,

    root: Dir = .{ .name = "" },
    cwd: ?*Dir = null,

    state: enum {
        ready,
        listing,
    } = .ready,

    pub fn init(allocator: Allocator) Self {
        return .{ .arena = std.heap.ArenaAllocator.init(allocator) };
    }

    pub fn deinit(self: *Self) void {
        self.arena.deinit();
    }

    pub fn getcwd(self: *Self) *Dir {
        return self.cwd orelse {
            const rp = &self.root;
            self.cwd = rp;
            return rp;
        };
    }

    pub fn eval(self: *Self, cur: *Parse.Cursor) !void {
        dispatch: while (true) {
            switch (self.state) {
                .ready => {
                    if (cur.have('$')) {
                        cur.expectStar(' ');
                        var cmd = try cur.expectToken(error.MissingCommandToken);
                        cur.expectStar(' ');
                        if (std.mem.eql(u8, cmd, "cd")) {
                            try self.chdir(cur.rem() orelse "");
                        } else if (std.mem.eql(u8, cmd, "ls")) {
                            try cur.expectEnd(error.UnexpectedArgument);
                            try self.list();
                        } else return error.NoSuchCommand;
                    } else return error.UnexpectedOutput;
                },

                .listing => {
                    if (cur.have('$')) {
                        cur.reset();
                        self.state = .ready;
                        continue :dispatch;
                    } else try self.parseListEntry(cur);
                },
            }

            break;
        }
    }

    fn chdir(self: *Self, path: []const u8) !void {
        var rem = path;
        if (rem.len > 0 and rem[0] == '/') {
            self.cwd = &self.root;
            rem = rem[1..];
        }
        if (rem.len == 0) {
            self.cwd = &self.root;
            return;
        }
        var parts = std.mem.split(u8, rem, "/");
        var cwd = self.getcwd();
        while (parts.next()) |name| {
            if (std.mem.eql(u8, name, "..")) {
                if (cwd.parent) |d|
                    cwd = d
                else
                    return if (cwd == &self.root) error.NoRootParent else error.NoDirParent;
            } else {
                var ent = cwd.getent(name) orelse return error.NoSuchEntry;
                switch (ent.*) {
                    .file => return error.NotDirectory,
                    .dir => |*d| cwd = d,
                }
            }
        }
        self.cwd = cwd;
    }

    fn list(self: *Self) !void {
        self.state = .listing;
    }

    fn parseListEntry(self: *Self, cur: *Parse.Cursor) !void {
        const allocator = self.arena.allocator();
        var item = try allocator.create(Dir.ListNode);
        item.next = null;

        var cwd = self.getcwd();

        if (cur.haveToken("dir")) {
            cur.expectStar(' ');
            const givenName = cur.rem() orelse return error.MissingDirName;
            item.ent = .{ .dir = .{
                .name = try allocator.dupe(u8, givenName),
                .parent = cwd,
            } };
        } else {
            const size = try cur.expectInt(usize, error.MissingFileSize);
            cur.expectStar(' ');
            const givenName = cur.rem() orelse return error.MissingFileName;
            item.ent = .{ .file = .{
                .name = try allocator.dupe(u8, givenName),
                .size = size,
            } };
        }

        try cwd.addent(item);
    }
};

const Parse = @import("./parse.zig");

fn run(
    allocator: Allocator,

    // TODO: better "any .reader()-able / any .writer()-able" interfacing
    input: anytype,
    output: anytype,
) !void {
    var lines = Parse.lineScanner(input.reader());
    var out = output.writer();

    var dev = Device.init(allocator);
    defer dev.deinit();

    while (try lines.next()) |*cur| {
        dev.eval(cur) catch |err| {
            return out.print("! {}\n- on line #{}: `{s}`\n", .{ err, cur.count, cur.buf });
        };
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
