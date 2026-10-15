# rhaelang

stack-based, concatenative language like Forth.

run with

```bash
go run . examples\rule110.rh
```

## lighting fast tutorial

**rhae** uses Reverse Polish Notation, that is, params go before the function call.

```go
// basic arithmetic
1 2 + println
13 5 >= println

// conditionals
10
dup 15 % 0 == if "FizzBuzz"
else dup 5 % 0 == if "Buzz"
else dup 3 % 0 == if "Fizz"
else dup
end end end
println

// loops
0 // sum
3 // starting from
loop dup 1000 < do
    dup                             // [sum, num, num]
    dup 5 % 0 ==                    // [sum, num, num, true/false]
    swap                            // [sum, num, true/false, num]
    3 % 0 ==                        // [sum, num, true/false, true/false]
    or if                           // [sum, num, true/false]
        dup                         // [sum, num, num]
    else
        0                           // [sum, num, num]
    end
    dup2over1 drop drop             // [num, num, sum]
    +                               // [num, num + sum]
    swap                            // [sum + num, num]
    1 +                             // [sum + num, num + 1]
end
drop    // drop the number which is 1000
println   // println the sum

// variables
3 start<            // store 3 into "start"
start> println      // load start onto the stack and print

// function
func sumTill do // ( int -- int )
    // = (n * (n + 1)) / 2
    dup 1 + * 2 /
end

// function - recursion works too
func fact do 
	dup 1 <= if
        1 *
    else
        dup 1 - fact *
    end
end

// list/array
list 			// list ( -- [T])
dup 11 add 		// add  ( [T] T -- )
dup 12 add
dup 13 add
dup 14 add
dup 15 add
dup 16 add

dup 0 get println	// get  ([T] int -- T)
dup 1 get println
dup 2 get println
dup 3 get println
dup 4 get println
dup 5 get println

dup 3 21 set	// set  ([T] int T -- )
dup 3 get println

dup size println	// size ( [T] -- int)

// some example string functions
"Hello world" 2 7 substr println

"Hello world" "llo" index println
"Hello world" "llo" contains println
"Hello world" "xyz" index println
"Hello world" "xyz" contains println
```