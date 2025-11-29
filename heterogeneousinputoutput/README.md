# Heterogeneous Input Output

- This exercise is about asynchronous input/output with mismatched (heterogeneous) inputs and outputs.
- **Level:** Intermediate
- **Topics:** Passing functions, goroutines, channels

--- 

## Funny Story

Roozbeh has recently moved from languages like JavaScript and Python to Go, and the concepts of goroutines, channels, 
and Go’s concurrency model still feel strange to him.
For this reason, he has asked you to implement some basic functions that behave like “Async I/O.”
In other words, he wants you to write a function that takes another function as input, 
runs it in the background (inside a separate goroutine), and later returns the result to him in some way.

### Hint

If you’re not familiar with concepts like: 
- _Future_ in JavaScript,
- _asyncio_ in Python,
- _ThreadPool_ or _async_ execution in other languages, it’s recommended to review these ideas first.
For example, this [article](https://www.baeldung.com/java-executor-service-tutorial) for JavaScript and this [article](https://realpython.com/async-io-python/) for Python can be helpful.

## Implementation & Explanation

You should complete a structure and several functions that work together to provide basic asynchronous behavior.
A Task is simply a function that takes no arguments and returns a string. Every async job in this project follows this signature.

```go
type FutureResult struct {
    Done       atomic.Bool
    ResultChan chan string
    // TODO
}

type Task func() string

func Async(t Task) *FutureResult {
}
func AsyncWithTimeout(t Task, timeout time.Duration) *FutureResult {
}

func (fResult *FutureResult) Await() string {
}

func CombineFutureResults(fResults ...*FutureResult) *FutureResult {
}
```

### `Async` Function

This is the most important function you need to implement.
As shown in the code, this function receives another function as input and returns an object of type `FutureResult`.
Even though `Async` itself runs synchronously and returns immediately, 
it must execute the provided function in the background, inside a separate goroutine.
In other words, `Async` should start the task asynchronously while the caller continues running without waiting for the task to finish.

### `Task` Type

The `Task` type is very simple — it is essentially a new name for a function that takes no input and returns a string.
All the operations Roozbeh wants to run in the background are functions of this type, so you don’t need to worry about passing arguments.
For simplicity, every task returns a string as its output.
These functions are usually time-consuming (for example, they may contain `time.Sleep`), and if executed sequentially in a single goroutine, they would take too long.
A simple example of a function that matches the `Task` type is:

```go
func simpleTask() string {
time.Sleep(1 * time.Second)
return "result"
}
```

### `FutureResult` Type
After a task is selected and its execution is started using the `Async` function, 
Roozbeh needs a way to access the task’s output and check its execution status.
For this purpose, the `FutureResult` struct provides a set of fields that help track and retrieve the result of an asynchronous task.

```go
type FutureResult struct {
    Done       atomic.Bool
    ResultChan chan string
    // TODO
}
```

1. As you can see, the first field is an `atomic.Bool`, which will be set to `true` once the task finishes. 
   - This type works differently from a normal boolean and provides thread-safe methods such as `Load` and `Store`.
   - You can check its [documentation](https://pkg.go.dev/sync/atomic#Bool) to learn more.
2. The second field is a `chan string` with a capacity of 1 (or more, as you will see later).
   - This channel is where the result of the task will be placed.
   - Once Roozbeh becomes more familiar with channels, he should be able to read directly from this channel.
   - When the task completes, a value will be sent into this channel, and Roozbeh can read it like this:

    ```go
    res := <-fResult.ResultChan
    ```

### `AsyncWithTimeout` Function

Some of Roozbeh’s tasks may take a long time to finish—for example, downloading a file over a very slow internet connection.
Roozbeh doesn’t want to wait forever, so you need to implement a function that accepts not only a `Task` but also a `timeout` of type `time.Duration`.
This `timeout` represents the maximum amount of time we are willing to wait for the task to complete.
More specifically, the `ResultChan` must receive a value no matter what, within the specified duration.
- If the task finishes before the `timeout` → the real result is sent to the channel
- If the `timeout` expires first → the string "timeout" is sent instead

```go
func AsyncWithTimeout(t Task, timeout time.Duration) *FutureResult {
}
```
#### Notes

- If a task times out, the value of `Done` does not matter, since Roozbeh has not defined a clear meaning for it in timeout scenarios.
- It is guaranteed that no normal Task will ever return the string "timeout", so this word is reserved exclusively to indicate a `timeout` event.
- Since it is not easy (or safe) to forcibly stop a goroutine in Go, you do not need to cancel the underlying task.
- The task may continue running in the background; this is acceptable and does not affect the tests.
- If the `timeout` duration is longer than the actual execution time of the task, the task result should simply be returned normally.
- The channel still has a capacity of 1, since it will contain either the real result or "timeout", **but never both**.

### `Await` Function

By now, it should be clear that the keyword `Async` doesn't mean much without an accompanying `Await`.
In this project, `Await` is a method defined on `*FutureResult` that Roozbeh can call to wait for the completion of his task without any additional complexity.
When `Await` is invoked, the caller blocks inside this method until the associated Task finishes and its result becomes available.

A correct usage example looks like this:

```go
fResult1 := Async(simpleTask)
fResult2 := Async(simpleTask)

res1 := fResult1.Await()
res2 := fResult2.Await()

assert.Equal(t, "result", res1)
assert.Equal(t, "result", res2)

```

**Note** that this example should not mislead you:
- Even though we call `Await` sequentially, both tasks are already running concurrently because each `Async` call started its task in a separate goroutine.
- So if `simpleTask` takes one second to complete, the entire code block above should still finish in about one second, **not two**.

### `CombineFutureResults` Function

The `CombineFutureResults` function helps Roozbeh merge several `FutureResult` objects into a single one.
The returned `FutureResult` should contain all results of the given tasks, in the same order that the input parameters were provided.
More precisely, imagine Roozbeh has launched multiple tasks and now wants to wait for every one of them.
Instead of calling `Await` separately on each task, he wants a single `FutureResult` whose output channel contains all results in sequence.
To achieve this, the output channel of the combined `FutureResult` must have a capacity equal to the number of input results.

```go
func CombineFutureResults(fResults ...*FutureResult) *FutureResult {
}
```

**Notes**
- It is guaranteed that the parameters passed to this function are not themselves the result of combining other FutureResults. They come directly from calls to `Async` or `AsyncWithTimeout`.
- Some of the combined results may come from timeout-enabled tasks.
- To simplify the problem, you can assume that `Await` will not be called on the object returned by this function; instead, the caller will read directly from the channel.

---

## Project Structure

```text
.
├── go.mod
├── go.sum
├── main.go                # your implementation
└── main_sample_test.go    # provided tests
```