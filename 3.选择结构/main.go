package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
)

type Object interface{}

// 全局变量/函数
var EnvMap = map[string]Object{
	"+": func(v []Object) Object {
		return v[0].(float64) + v[1].(float64)
	},
	"-": func(v []Object) Object {
		return v[0].(float64) - v[1].(float64)
	},
	"*": func(v []Object) Object {
		return v[0].(float64) * v[1].(float64)
	},
	"/": func(v []Object) Object {
		return v[0].(float64) / v[1].(float64)
	},
	"^": func(v []Object) Object { // 指数
		return math.Pow(v[0].(float64), v[1].(float64))
	},
	">": func(v []Object) Object { // 指数
		return v[0].(float64) > v[1].(float64)
	},
	">=": func(v []Object) Object { // 指数
		return v[0].(float64) >= v[1].(float64)
	},

	"<": func(v []Object) Object { // 指数
		return v[0].(float64) < v[1].(float64)
	},
	"<=": func(v []Object) Object { // 指数
		return v[0].(float64) <= v[1].(float64)
	},
	"==": func(v []Object) Object { // 指数
		return v[0].(float64) == v[1].(float64)
	},
	"!=": func(v []Object) Object { // 指数
		return v[0].(float64) != v[1].(float64)
	},
	"sin": func(v []Object) Object {
		return math.Sin(v[0].(float64))
	},
	"cos": func(v []Object) Object {
		return math.Cos(v[0].(float64))
	},
	"tan": func(v []Object) Object {
		return math.Tan(v[0].(float64))
	},
	"print": func(v []Object) Object {
		fmt.Println(v)
		return nil
	},
	// 其余可自行添加
}

func IsAbc(v byte) bool { // 字母判断
	if (v >= 'a' && v <= 'z') || (v >= 'A' && v <= 'Z') {
		return true
	}
	return false
}

func IsSym(v byte) bool { // 符号判断
	syms := "@#$%^&*()-+=/?|!`~<>"
	for i := 0; i < len(syms); i++ {
		if syms[i] == v {
			return true
		}
	}
	return false
}

func IsNum(v byte) bool { // 数字判断
	if v >= '0' && v <= '9' {
		return true
	}
	return false
}

type Code struct {
	src    string // 源码
	pos    int    // 当前指针位置
	tokens []Object
}

func (self *Code) Init(src string) { // 初始化
	self.src = src
	self.pos = 0
	self.Next() // 初始化就读一个token
}
func (self *Code) Peek() Object { // 返回当前token
	if len(self.tokens) > 0 {
		return self.tokens[len(self.tokens)-1]
	}
	return nil
}
func (self *Code) Next() Object { // 下一个token
	if self.pos >= len(self.src) {
		return nil
	}
	var tmp []byte
	var tk Object
	var i = self.pos

	switch {
	case IsNum(self.src[i]) || (self.src[i] == '-' && IsNum(self.src[i+1])): // 读取数字
		for i < len(self.src) && (IsNum(self.src[i]) || self.src[i] == '.' || self.src[i] == '-') {
			tmp = append(tmp, self.src[i])
			i++
		}
		self.pos = i
		tk, _ = strconv.ParseFloat(string(tmp), 64)
	case IsSym(self.src[i]): // 读取字符
		if self.src[i] == '(' {
			tmp = append(tmp, self.src[i])
			i++
		} else {
			for i < len(self.src) && IsSym(self.src[i]) && self.src[i] != '(' {
				tmp = append(tmp, self.src[i])
				i++
			}
		}
		self.pos = i
		tk = string(tmp)
	case IsAbc(self.src[i]): // 读取字母（变量）
		for i < len(self.src) && IsAbc(self.src[i]) {
			tmp = append(tmp, self.src[i])
			i++
		}
		self.pos = i
		tk = string(tmp)
	default:
		self.pos++
		tk = self.Next()
	}
	self.tokens = append(self.tokens, tk)
	return tk
}
func (self *Code) read_list() []Object { // 读列表
	var lt []Object
	v := self.Peek()
	for v != ")" && v != nil {
		if v == "(" {
			self.Next()
			lt = append(lt, self.read_list())
		} else {
			lt = append(lt, v)
		}
		v = self.Next()
	}
	return lt
}

func (self *Code) Read_Root() Object { // 根节点开始解析
	v := self.Peek()
	if v == "(" {
		self.Next()
		return self.read_list()
	} else {
		return v
	}
}
func Apply(v []Object, env *map[string]Object, fn func(Object, *map[string]Object) Object) []Object { // 将函数fn 应用到列表每一项
	var res []Object
	for _, j := range v {
		res = append(res, fn(j, env))
	}
	return res
}
func Eval(tree Object, env *map[string]Object) Object { // 计算表达式
	/**
	  tree token对象
	  env 环境（变量、函数）
	*/
	switch tree.(type) {
	case []Object:
		v, _ := tree.([]Object)
		if len(v) < 2 {
			return nil
		}
		// 取出操作符及其对应的函数
		op := v[0].(string)
		switch op {
		case "set": // 设置变量值(set a 12)或者(set f (+ 1 2))
			key := v[1].(string)
			val := Eval(v[2], env)
			(*env)[key] = val
			return val
		case "if": // 判断语句(if true 3 5)
			du, ok := Eval(v[1], env).(bool)
			if ok {
				if du {
					return Eval(v[2], env)
				} else {
					return Eval(v[3], env)
				}
			}
		default:
			f, ok := (*env)[op].(func([]Object) Object)
			if ok {
				return f(Apply(v[1:], env, Eval))
			}
		}

	case Object:
		switch tree.(type) {
		case string:
			if k, ok := (*env)[tree.(string)]; ok {
				//存在变量
				return k
			}
		}
		return tree
	}
	return nil
}
func main() {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("User>>")
		cmds, _, _ := reader.ReadLine()
		cmd := string(cmds)
		if cmd == "exit" {
			return
		} else {
			c := Code{}
			c.Init(cmd)
			s := c.Read_Root()
			// fmt.Println(s)
			fmt.Println(Eval(s, &EnvMap))
		}

	}

}
