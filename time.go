package main

import (
    "time"
    "fmt"
)

func now() string {
    return fmt.Sprintf("%s", time.Now())
}
