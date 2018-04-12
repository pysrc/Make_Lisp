# 用Go实现仿Lisp解释器

觉得自己写一个语言很有成就感，其实也没什么特别牛逼的东西，就是实现一些概念，把过程中的坑点记录下来了，项目参考 https://github.com/kanaka/mal 的实现，由于英文不好，一边做一边看这位兄弟 https://github.com/Windfarer/mal-zh 的中文翻译，翻译得很好。没有实现Lisp的全部功能，只实现了基本的运算功能

## 示例

```lisp
// fib.txt
S:
(fn // 斐波拉契数列递归式
    fib 
    [n] 
    {
        (if 
            (<= n 2) 
            {
                (ret 1)
            } 
            {
                (ret 
                    (+ 
                        (fib (- n 1))
                        (fib (- n 2))
                    )
                )
            }
        )
    }
)
:E
(out (fib 6))
```

`main.exe fib.txt`

[更多示例请看这里](/一些示例)