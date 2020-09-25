# generator
Go Code Generator

## Usage

### Getter

例如下面的结构体

```go
package mystruct

//go:generate god getter -t SomeStruct
type SomeStruct struct {
   // private field
   field1 string    // 会产生 `Field1` 方法

   // public field
   Field2 string    // 不会产生新方法

   // both private and public field
   field3 string    // 不会产生新方法并会导致警告
   Field3 string

   // 已经自定义了 Getter Field4
   field4 string   // 不会产生新方法

   // ignore field
   field5 string  `getter:"ignore"` // 不会产生新方法
}

func (ss *SomeStruct) Field4() string {
    return "Field4: " + ss.field4
}
```

生成遵从以下规则

1. 只有 private field 会产生 Getter
2. 如果指定了 `getter:"ignore"` 则不会产生 Getter
3. 如果已经有了自定义的 Getter 则不会额外产生新的 Getter
4. 如果有同名（指的是只有首字母大小写不同）属性则不会产生 Getter 并且会给出警告
5. 满足上述所有条件后会生成 Getter，例如 `field1` 会导致结构体增加 `Field1` 方法并返回 `field1` 所对应的值

## License

This software is released under the Apache-2.0 license.

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FImSingee%2Fgenerator.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FImSingee%2Fgenerator?ref=badge_large)
