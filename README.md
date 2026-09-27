Go Fuzz Testing - The Basics
============================

![Andrei Serban](data:image/jpeg;base64,/9j/2wBDABALDA4MChAODQ4SERATGCgaGBYWGDEjJR0oOjM9PDkzODdASFxOQERXRTc4UG1RV19iZ2hnPk1xeXBkeFxlZ2P/2wBDARESEhgVGC8aGi9jQjhCY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2P/wgARCAAUABQDASIAAhEBAxEB/8QAGQABAQEAAwAAAAAAAAAAAAAAAAMBAgQF/8QAFwEAAwEAAAAAAAAAAAAAAAAAAAIDAf/aAAwDAQACEAMQAAABtKfGL9xNh4Wl00B//8QAGxAAAwEAAwEAAAAAAAAAAAAAAQIDABAREjH/2gAIAQEAAQUCY+QrNxSwUm88G7H1sKMu/8QAFBEBAAAAAAAAAAAAAAAAAAAAIP/aAAgBAwEBPwEf/8QAFxEBAAMAAAAAAAAAAAAAAAAAAQAQEf/aAAgBAgEBPwEJl//EABoQAAMAAwEAAAAAAAAAAAAAAAABERAhIlH/2gAIAQEABj8Cp2lvx4jFCodxpn//xAAbEAADAQEAAwAAAAAAAAAAAAAAAREhMUFRcf/aAAgBAQABPyGzFhSki+NCl8b6FNL8Qpeifkbu6bIKJJH/2gAMAwEAAgADAAAAEM/wvf/EABgRAQEAAwAAAAAAAAAAAAAAABEAARAx/9oACAEDAQE/EHB2dN//xAAYEQADAQEAAAAAAAAAAAAAAAAAARExEP/aAAgBAgEBPxBszhYRH//EAB4QAAICAgIDAAAAAAAAAAAAAAERACExURCRQWGB/9oACAEBAAE/EBtyGnmLMQnWb1wZ2a1D7gUAsbOP2aHUOCMYPMPZiCT9iahuYxqf/9k=)![Andrei Serban](/web/20220329152653im_/https://blog.fuzzbuzz.io/static/3d66614e99116903d01446c63cb9d705/78d94/profile-pic.jpg)

![Andrei Serban](/web/20220329152653im_/https://blog.fuzzbuzz.io/static/3d66614e99116903d01446c63cb9d705/78d94/profile-pic.jpg)

[Andrei Serban](https://web.archive.org/web/20220329152653/https://www.linkedin.com/in/andrei-serban)

March 29, 2022 ~ 12 min read

Tutorial

How many bugs can you find in this function? It looks simple enough - take a string, and overwrite the first `n` characters with a new user-defined character. For example, if we ran `OverwriteString("Hello, World!", "A", 5)`, we would expect the output: `"AAAAA, World!"`.

    // overwrite_string.go
    
    // OverwriteString overwrites the first 'n' characters in a string with
    // the rune 'value'
    func OverwriteString(str string, value rune, n int) string {
    	// If asked to overwrite more than the entire string then no need to loop,
    	// just return string length * the rune
    	if n > len(str) {
    		return strings.Repeat(string(value), len(str))
    	}
    
    	result := []rune(str)
    	for i := 0; i <= n; i++ {
    		result[i] = value
    	}
    	return string(result)
    }

In the time it takes to give the code a quick visual once-over, [Go’s new fuzz testing tool](https://web.archive.org/web/20220329152653/https://go.dev/doc/fuzz/) can run over 5 million procedurally generated inputs through the function, and in this case find an input that causes an out-of-bounds array access in just _one second_.

It turns out running the function with this set of arguments: `OverwriteString("0000", rune('A'), 4)` causes a panic:

    --- FAIL: FuzzBasicOverwriteString (0.05s)
        --- FAIL: FuzzBasicOverwriteString (0.00s)
            testing.go:1349: panic: runtime error: index out of range [4] with length 4
                goroutine 96 [running]:
                runtime/debug.Stack()
                	/home/everest/sdk/gotip/src/runtime/debug/stack.go:24 +0x90
                testing.tRunner.func1()
               ...<snip> 
        
        Failing input written to testdata/fuzz/FuzzBasicOverwriteString/2bac7bdf139ad0b2de37275db2a606ecb335bd344500173b451e9dfc3658c12f
        To re-run:
        go test -run=FuzzBasicOverwriteString/2bac7bdf139ad0b2de37275db2a606ecb335bd344500173b451e9dfc3658c12f

Fuzz testing (or fuzzing) is a powerful testing technique that’s great at discovering bugs and vulnerabilities that developers typically miss, and has a strong track record of discovering [hundreds of critical bugs in open-source Go code](https://web.archive.org/web/20220329152653/https://github.com/dvyukov/go-fuzz#trophies).

Extend our little example problem to a thousand-line codepath in a critical application, and a fuzzer churning through billions of inputs can take only a few minutes to discover subtle bugs that would otherwise take days to unravel in production. What follows is a first look at how to ramp up on Go’s latest testing tool and start discovering your own bugs as quickly as possible.

Getting Started
---------------

We’ll be using the new fuzz testing features in Go 1.18, so before we get started, make sure that your `$ go version` is at least 1.18. Head to [https://go.dev/dl](https://web.archive.org/web/20220329152653/https://go.dev/dl/) to update your `go` tool if you’re on an earlier version.

If you’d like to follow along, you can find the code for this post at [github.com/fuzzbuzz/go-fuzzing-tutorial](https://web.archive.org/web/20220329152653/https://github.com/fuzzbuzz/go-fuzzing-tutorial). For the remainder of this tutorial, you can assume all commands you see are run from inside the `introduction` directory.

This is the most basic incarnation of a fuzz test:

    // overwrite_string_test.go
    
    func FuzzBasicOverwriteString(f *testing.F) {
    	f.Fuzz(func(t *testing.T, str string, value rune, n int) {
    		OverwriteString(str, value, n)
    	})
    }

Contrary to a unit test, which expects a specific behavior from a fixed input, a fuzz test runs thousands of procedurally generated inputs through the function it’s testing, without the developer needing to manually come up with inputs. In this specific case, we passed the function we wish to test to `f.Fuzz`, so the fuzzer will generate a new `string`, `rune`, and `int` to fill the arguments for each test iteration.

By default, fuzzing will detect crashes, hangs, and extreme memory usage, so even without writing any assertions we’ve already built a useful robustness test for our function.

To run this test, type:

    go test -fuzz FuzzBasicOverwriteString

Within about a second, you should see the test exit with an error signature similar to this:

    --- FAIL: FuzzBasicOverwriteString (0.05s)
        --- FAIL: FuzzBasicOverwriteString (0.00s)
            testing.go:1349: panic: runtime error: index out of range [4] with length 4
    ...SNIP
        Failing input written to testdata/fuzz/FuzzBasicOverwriteString/2bac7bdf139ad0b2de37275db2a606ecb335bd344500173b451e9dfc3658c12f
        To re-run:
        go test -run=FuzzBasicOverwriteString/2bac7bdf139ad0b2de37275db2a606ecb335bd344500173b451e9dfc3658c12f

The fuzzer provides the location of the specific input that caused the problem, inside the `testdata/fuzz/FuzzBasicOverwriteString` directory. If you open up this file, you can take a look at the actual values that caused our function to panic:

    go test fuzz v1
    string("00")
    rune('A')
    int(2)

Now that we’ve found a bug, we can go into our code and fix the problem. Taking a look at the line of code that actually caused the panic (`overwrite_string.go:16`), it seems the code tried to access index 4 of a string with length 4, which caused an array index out of range error. You can fix the bug by changing the `if n > len(str)` check to test for greater than or equal to instead: `if n >= len(str)`:

    // overwrite_string.go
    
    // OverwriteString overwrites the first 'n' characters in a string with
    // the rune 'value'
    func OverwriteString(str string, value rune, n int) string {
    	// If asked to overwrite more than the entire string then no need to loop,
    	// just return string length * the rune
    	if n >= len(str) {
    		return strings.Repeat(string(value), len(str))
    	}
    
    	result := []rune(str)
    	for i := 0; i <= n; i++ {
    		result[i] = value
    	}
    	return string(result)
    }

This will make sure the loop is only entered if `n` is at least 1 less than the string length. We _could_ fix the for loop’s bounds as well, but that hides a more interesting bug later on, so for now, we can pretend it was overlooked.

Confirm this fixes the bug by re-running the crashing test case using the command the fuzzer provided in the output:

    $ go test -v -count=1 -run=FuzzBasicOverwriteString/2bac7bdf139ad0b2de37275db2a606ecb335bd344500173b451e9dfc3658c12f
    
    === RUN   FuzzBasicOverwriteString
    === RUN   FuzzBasicOverwriteString/2bac7bdf139ad0b2de37275db2a606ecb335bd344500173b451e9dfc3658c12f
    --- PASS: FuzzBasicOverwriteString (0.00s)
        --- PASS: FuzzBasicOverwriteString/2bac7bdf139ad0b2de37275db2a606ecb335bd344500173b451e9dfc3658c12f (0.00s)
    PASS
    ok  	github.com/fuzzbuzz/go-fuzzing-tutorial/introduction	0.001s

Go’s fuzzer will automatically run each input in the `testdata` directory as a unit test any time `go test` is run (these inputs are collectively called “seeds”). Committing the `testdata` directory to version control will save this input as a permanent regression test to ensure the bug is never reintroduced.

A Happy Accident
----------------

Now, I will fully admit that as I was writing this, I expected this change to satisfy the basic fuzz test, but if you re-run the fuzzer after this change you’ll notice that an entirely new bug manifests:

    $ go test -fuzz FuzzBasicOverwriteString
    fuzz: elapsed: 0s, gathering baseline coverage: 0/17 completed
    fuzz: elapsed: 0s, gathering baseline coverage: 17/17 completed, now fuzzing with 8 workers
    fuzz: minimizing 177-byte failing input file
    fuzz: elapsed: 0s, minimizing
    --- FAIL: FuzzBasicOverwriteString (0.17s)
        --- FAIL: FuzzBasicOverwriteString (0.00s)
            testing.go:1349: panic: runtime error: index out of range [60] with length 60
                goroutine 2911 [running]:
                runtime/debug.Stack()
                	/home/everest/sdk/gotip/src/runtime/debug/stack.go:24 +0x90
                testing.tRunner.func1()
                	/home/everest/sdk/gotip/src/testing/testing.go:1349 +0x1f2
                panic({0x5b3700, 0xc00289c798})
                	/home/everest/sdk/gotip/src/runtime/panic.go:838 +0x207
                github.com/fuzzbuzz/go-fuzzing-tutorial/introduction.OverwriteString({0xc00288ef00, 0x3d}, 0x83, 0x3c)
                	/home/everest/src/fuzzbuzz/go-fuzzing-tutorial/introduction/overwrite_string.go:20 +0x270
                github.com/fuzzbuzz/go-fuzzing-tutorial/introduction.FuzzBasicOverwriteString.func1(0x5?, {0xc00288ef00?, 0x0?}, 0x0?, 0x0?)
                	/home/everest/src/fuzzbuzz/go-fuzzing-tutorial/introduction/overwrite_string_test.go:24 +0x38
                reflect.Value.call({0x598d60?, 0x5cfb58?, 0x13?}, {0x5c179f, 0x4}, {0xc0028c2de0, 0x4, 0x4?})
                	/home/everest/sdk/gotip/src/reflect/value.go:556 +0x845
                reflect.Value.Call({0x598d60?, 0x5cfb58?, 0x514?}, {0xc0028c2de0, 0x4, 0x4})
                	/home/everest/sdk/gotip/src/reflect/value.go:339 +0xbf
                testing.(*F).Fuzz.func1.1(0x0?)
                	/home/everest/sdk/gotip/src/testing/fuzz.go:337 +0x231
                testing.tRunner(0xc0028e7380, 0xc0028ec5a0)
                	/home/everest/sdk/gotip/src/testing/testing.go:1439 +0x102
                created by testing.(*F).Fuzz.func1
                	/home/everest/sdk/gotip/src/testing/fuzz.go:324 +0x5b8
                
        
        Failing input written to testdata/fuzz/FuzzBasicOverwriteString/2ee896e38866e089811eeece13f9919795072e6cc05ee9f782d68d1663d204c7
        To re-run:
        go test -run=FuzzBasicOverwriteString/2ee896e38866e089811eeece13f9919795072e6cc05ee9f782d68d1663d204c7
    FAIL
    exit status 1
    FAIL	github.com/fuzzbuzz/go-fuzzing-tutorial/introduction	0.174s

The actual input the fuzzer generated may look different in your case, but this is the test case that was generated for me:

    go test fuzz v1
    string("000000000000000000000000000000Ö00000000000000000000000000000")
    rune('\u0083')
    int(60)

At first glance, this bug looks almost exactly like the one you just fixed. Trying to access index 60 of a 60-character long string shouldn’t be possible, since the function would return at the initial if statement. But herein lies the power of fuzzing - it uncovers edge cases developers haven’t considered, and this is in fact an entirely separate bug.

If you inspect your crashing input you’ll likely notice, like mine, that one of the characters is a Unicode character. That is, it is represented by more than one byte. In my case, it’s `Ö`. Sure, this input string is 60 characters long, but it’s _61 bytes long_. And it turns out that in Go, taking the `len` of a string returns the number of _bytes_ in the string, not the number of characters (or runes, to use Go’s terminology).

This is easy to check for yourself. If you run the following snippet of Go code:

    str := "Ö"
    runeArray := []rune(str)
    fmt.Println("Str len:", len(str), "Rune array len:", len(runeArray))

You’ll see the following output:

    Str len: 2 Rune array len: 1

With that information about Go’s string implementation top-of-mind, rewrite the if statement again, from `if n >= len(str)` to `if n >= utf8.RuneCountInString(str)`. We want to compare apples to apples, or in this case, quantities of characters to characters, not bytes:

    // overwrite_string.go
    
    // OverwriteString overwrites the first 'n' characters in a string with
    // the rune 'value'
    func OverwriteString(str string, value rune, n int) string {
    	// If asked to overwrite more than the entire string then no need to loop,
    	// just return string length * the rune
    	if n >= utf8.RuneCountInString(str) {
    		return strings.Repeat(string(value), len(str))
    	}
    
    	result := []rune(str)
    	for i := 0; i <= n; i++ {
    		result[i] = value
    	}
    	return string(result)
    }

Run the fuzz test once more, and watch it churn away, attempting to find another input to crash our function:

    go test -fuzz FuzzBasicOverwriteString

You _should_ let this run for a while to make sure that there aren’t any other bugs lurking around, but you can at least be confident that the function doesn’t crash on the most basic of inputs. You can press `ctrl-C` to stop the fuzzer.

Functional Bugs
---------------

Up to this point, we’ve found bugs that result in crashes. Denial of service is a big deal, but all we know is that this function doesn’t crash on unexpected inputs. It’s also important to test the _correctness_ of the function. There are a number of ways to approach this, but with fuzz testing, it’s best to try to think of a _an invariant (or property)_ that will always hold true for your code.

An example of an invariant for `OverwriteString` is that the function should never fill _more_ characters than the number it’s asked to. More concretely, if asked to overwrite “Hello, world!” with 5 “A” characters, it should be possible to check that the remaining characters in the string are still “, world!”.

This can be generalized with the following test:

    // overwrite_string_test.go
    
    func FuzzOverwriteStringSuffix(f *testing.F) {
    	f.Add("Hello, world!", 'A', 15)
    
    	f.Fuzz(func(t *testing.T, str string, value rune, n int) {
    		result := OverwriteString(str, value, n)
    		if n > 0 && n < utf8.RuneCountInString(str) {
    			// If we modified characters [0:n], then characters [n:] should stay the same
    			resultSuffix := string([]rune(result)[n:])
    			strSuffix := string([]rune(str)[:])
    			if resultSuffix != strSuffix {
    				t.Fatalf("OverwriteString modified too many characters! Expected %s, got %s.", strSuffix, resultSuffix)
    			}
    		}
    	})
    }

Running the test will produce another bug:

    $ go test -fuzz FuzzOverwriteStringSuffix
    
    fuzz: elapsed: 0s, gathering baseline coverage: 0/54 completed
    fuzz: minimizing 66-byte failing input file
    fuzz: elapsed: 0s, gathering baseline coverage: 8/54 completed
    --- FAIL: FuzzOverwriteStringSuffix (0.03s)
        --- FAIL: FuzzOverwriteStringSuffix (0.00s)
            overwrite_string_test.go:38: OverwriteString modified too many characters! Expected 0, got A.
        
        Failing input written to testdata/fuzz/FuzzOverwriteStringSuffix/148139e8febb077401421c031a9bd3c3315179c5a66c90349102d223b451ec02
        To re-run:
        go test -run=FuzzOverwriteStringSuffix/148139e8febb077401421c031a9bd3c3315179c5a66c90349102d223b451ec02
    FAIL
    exit status 1
    FAIL	github.com/fuzzbuzz/go-fuzzing-tutorial/introduction	0.031s

Note that this time it’s not a panic, but instead a message that looks very similar to a unit test failure. It turns out there’s been a functional bug in this code the whole time.

Since it checks if the loop index is `<=` on line 20, it’s been filling in one too many characters without realizing it. Changing the loop condition from `i <= n` to `i < n` solves this problem.

The final `OverwriteString` function should look like this:

    // overwrite_string.go
    
    // OverwriteString overwrites the first 'n' characters in a string with
    // the rune 'value'
    func OverwriteString(str string, value rune, n int) string {
    	// If asked for more no need to loop, just return
    	// string length * the rune
    	if n >= utf8.RuneCountInString(str) {
    		return strings.Repeat(string(value), len(str))
    	}
    
    	result := []rune(str)
    	fmt.Println("HELLO", str, len(str), len(result), n)
    	for i := 0; i < n; i++ {
    		result[i] = value
    	}
    	return string(result)
    }

If you run the fuzzer once more, you should be greeted with the fuzzer reliably running thousands of inputs per second without discovering another bug.

Ideally this fuzz test would run for at least a few minutes to increase confidence in this code’s correctness (especially if it were testing more than a 10 line function).

The bugs in this post were found in a few seconds, but some bugs can take hours or days of fuzzing as the fuzzer needs time to explore the entire state space of the software under test. We’ll cover the art of continuous fuzzing at scale in a follow-up post in the next few weeks.

Wrapping up
-----------

This was just a brief introduction to fuzz testing with Go. There are more advanced topics we could get into, but the examples discussed today are really all you need to start adding fuzz tests to your own projects.

In the coming weeks we’ll dive deeper into the types of bugs you can find, investigate some real-world fuzzing bugs, and discuss how to automate your fuzzing so that CI finds bugs while you’re asleep. You can subscribe to our updates below to get notified when the next part comes out.

In the meantime, if you’d like to take this tutorial a little further, consider that we’ve only written a fuzz test that checks the part of the string we _shouldn’t_ change. How would you write a fuzz test that confirms that we’ve actually modified the desired characters correctly?
