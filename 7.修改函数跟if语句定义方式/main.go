package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
)

var DEBUG = false

func Debug(v ...interface{}) {
	if DEBUG {
		fmt.Println(v...)
	}
}

type Object interface{}

type EnvType struct { // 环境，列表最后一个*map为内环境，其余为外环境
	Val [](*map[string]Object)
}

func (self *EnvType) Copy() *EnvType {
	var env EnvType
	inner_env := make(map[string]Object) // 内环境
	for _, v := range self.Val {
		env.Val = append(env.Val, v)
	}
	env.Val = append(env.Val, &inner_env)
	return &env
}
func (self *EnvType) Set(key string, val Object) { // 设置key-val
	for i := len(self.Val) - 1; i >= 0; i-- { // 从内环境向外查找
		if _, ok := (*self.Val[i])[key]; ok {
			(*self.Val[i])[key] = val
			return
		}
	}
	// 找不到就在内环境设置
	(*self.Val[len(self.Val)-1])[key] = val
}
func (self *EnvType) Get(key string) Object { // 获取value
	for i := len(self.Val) - 1; i >= 0; i-- { // 从内环境向外查找
		if v, ok := (*self.Val[i])[key]; ok {
			return v
		}
	}
	return nil
}
func (self *EnvType) Find(key string) bool { // 判断key是否存在
	for i := len(self.Val) - 1; i >= 0; i-- { // 从内环境向外查找
		if _, ok := (*self.Val[i])[key]; ok {
			return true
		}
	}
	return false
}

type Fn struct { // 函数结构
	Name string   // 函数名
	Args []Object // 形参
	Body Object   // 函数体
	Env  *EnvType
}

// 全局变量/函数
var EnvMap = map[string]Object{
	"+": func(v []Object) Object { // 连加支持
		var res float64 = 0
		for _, i := range v {
			res += i.(float64)
		}
		return res
	},
	"-": func(v []Object) Object { // 连减支持
		var res float64 = v[0].(float64)
		for i := 1; i < len(v); i++ {
			res -= v[i].(float64)
		}
		return res
	},
	"*": func(v []Object) Object { // 连乘支持
		var res float64 = 1
		for _, i := range v {
			res *= i.(float64)
		}
		return res
	},
	"/": func(v []Object) Object { // 连除支持
		var res float64 = v[0].(float64)
		for i := 1; i < len(v); i++ {
			res /= v[i].(float64)
		}
		return res
	},
	"^": func(v []Object) Object { // 指数
		return math.Pow(v[0].(float64), v[1].(float64))
	},
	">": func(v []Object) Object {
		return v[0].(float64) > v[1].(float64)
	},
	">=": func(v []Object) Object {
		return v[0].(float64) >= v[1].(float64)
	},

	"<": func(v []Object) Object {
		return v[0].(float64) < v[1].(float64)
	},
	"<=": func(v []Object) Object {
		return v[0].(float64) <= v[1].(float64)
	},
	"==": func(v []Object) Object {
		return v[0].(float64) == v[1].(float64)
	},
	"!=": func(v []Object) Object {
		return v[0].(float64) != v[1].(float64)
	},
	"&&": func(v []Object) Object {
		return v[0].(bool) && v[1].(bool)
	},
	"||": func(v []Object) Object {
		return v[0].(bool) || v[1].(bool)
	},
	"!": func(v []Object) Object {
		return !v[0].(bool)
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
	"mod": func(v []Object) Object { // 取余
		return math.Mod(v[0].(float64), v[1].(float64))
	},
	"%": func(v []Object) Object {
		return math.Mod(v[0].(float64), v[1].(float64))
	},
	"exp": func(v []Object) Object {
		return math.Exp(v[0].(float64))
	},
	"log": func(v []Object) Object { // 以 e为底
		return math.Log(v[0].(float64))
	},
	"out": func(v []Object) Object { // 输出函数
		fmt.Println(v)
		return nil
	},
	"ret": func(v []Object) Object {
		return v[0]
	},
	// 其余可自行添加
}
var EnvG = EnvType{[]*map[string]Object{&EnvMap}}

func IsSym(v byte) bool { // 括号判断
	syms := "(){}[]"
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

/**词法分析开始**/
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
	case IsSym(self.src[i]): // 括号判断
		tmp = append(tmp, self.src[i])
		i++
		self.pos = i
		tk = string(tmp)
	case self.src[i] == ' ': // 空格
		self.pos++
		tk = self.Next()
	default: // 变量||函数
		for i < len(self.src) && self.src[i] != ' ' && !IsSym(self.src[i]) {
			tmp = append(tmp, self.src[i])
			i++
		}
		self.pos = i
		tk = string(tmp)
	}
	self.tokens = append(self.tokens, tk)
	return tk
}

func (self *Code) read_list() []Object { // 读列表
	var lt []Object
	v := self.Peek()
	for v != ")" && v != "]" && v != "}" && v != nil {
		if v == "(" || v == "[" || v == "{" {
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
	if v == "(" || v == "[" || v == "{" { // 读列表
		self.Next()
		return self.read_list()
	} else {
		return v
	}
}

/**词法分析结束**/

/**计算开始**/
func Apply(v []Object, env *EnvType, fn func(Object, *EnvType) Object) []Object { // 将函数fn 应用到列表每一项
	var res []Object
	for _, j := range v {
		res = append(res, fn(Eval(j, env), env))
	}
	return res
}
func Eval(tree Object, env *EnvType) Object { // 计算表达式
	/**
	  tree token对象
	  env 环境（变量、函数）
	*/
	Debug("AST:", tree)
	switch tree.(type) {
	case []Object:
		v, _ := tree.([]Object)
		// fmt.Println("switch:", len(v))
		if len(v) < 2 {
			return nil
		}
		// 取出操作符及其对应的函数
		op := v[0].(string)
		// fmt.Println("op:", op)
		switch op {
		case "set", "=": // 设置变量值(set a 12)或者(set f (+ 1 2))
			val := Eval(v[2], env)
			env.Set(v[1].(string), val)
			return val
		case "if": // 判断语句(if (bool expr) {expr1 expr2 ...} {expr3 expr4 ...})
			du, ok := Eval(v[1], env).(bool)
			if_env := env.Copy()
			if ok {
				if du {
					switch v[2].(type) {
					case []Object:
						exprs := v[2].([]Object)
						for i, expr := range exprs {
							if i == len(exprs)-1 {
								return Eval(expr, if_env) // 最后一个表达式值作为返回值
							} else {
								Eval(expr, if_env) // 中间表达式值不返回
							}
						}
					case Object: // 暂不处理单个元素
						fmt.Println("if 结构错误！正确格式为：(if (bool expr) {expr1 expr2 ...} {expr3 expr4 ...})")
						return nil
					}

				} else if len(v) == 4 { // 有else时
					switch v[3].(type) {
					case []Object:
						exprs := v[3].([]Object)
						for i, expr := range exprs {
							if i == len(exprs)-1 {
								return Eval(expr, if_env) // 最后一个表达式值作为返回值
							} else {
								Eval(expr, if_env) // 中间表达式值不返回
							}
						}
					case Object:
						fmt.Println("if 结构错误！正确格式为：(if (bool expr) {expr1 expr2 ...} {expr3 expr4 ...})")
						return nil
					}
				}
			}
		case "fn": // 函数定义 (fn fn_name [x y ... ] {expr1 expr2 ...})
			var fn Fn
			fn.Env = env.Copy()
			fn.Name = v[1].(string)
			fn.Args = v[2].([]Object)
			fn.Body = v[3].([]Object)
			env.Set(fn.Name, fn)    // 向上一层环境中加入函数
			fn.Env.Set(fn.Name, fn) // 要想实现递归,就应当在自己的环境中找到自己,这是必须的
			return fn
		case "for": // 循环语句(for (bool expr) {(expr1) (expr2) (expr3) ...})
			exprs := v[2].([]Object) // 循环体
			for_env := env.Copy()
			for Eval(v[1], for_env).(bool) { // v[1] 是循环判断结构
				for _, expr := range exprs { // 执行循环体
					Eval(expr, for_env)
				}
			}
		default:
			if env.Find(op) {
				f := env.Get(op)
				switch f.(type) {
				case func([]Object) Object: // 系统函数
					fc := f.(func([]Object) Object)
					return fc(Apply(v[1:], env, Eval))
				case Fn: // 自定义函数 (fn_name args1 args2 ...)
					// fmt.Println("自定义函数:", op)
					fc := f.(Fn)
					fenv := fc.Env.Copy()
					// 取传入函数的参数(可能是表达式)
					args := v[1:]
					for i, j := range args { // 将传递的参数加入函数环境
						fenv.Set(fc.Args[i].(string), Eval(j, env))
					}
					body := fc.Body.([]Object)
					for i, expr := range body { // 函数最后一个语句作为返回值
						if i == len(body)-1 {
							return Eval(expr, fenv)
						} else {
							Eval(expr, fenv)
						}
					}
				}

			}
		}

	case Object:
		switch tree.(type) {
		case string:
			str := tree.(string)
			b, err := strconv.ParseBool(str) // 先转换为bool
			if err == nil {
				return b
			} else if env.Find(str) {
				//存在变量
				return env.Get(str)
			}
		}
		return tree
	}
	return nil
}
func FindExpr(s string) string { // 找出字符串中的语句
	var x, y int
	for i := 0; i < len(s); i++ {
		if s[i] == '(' {
			x = i
			break
		}
	}
	for i := len(s) - 1; i > 0; i-- {
		if s[i] == ')' {
			y = i
			break
		}
	}
	if x < y { // 找到
		return s[x : y+1]
	} else {
		return "" // 找不到
	}
}
func ExeFile(cmd string) { // 执行文件
	cmds := strings.Split(cmd, "\n") // 分离语句
	for i := 0; i < len(cmds); i++ {
		sr := strings.Trim(cmds[i], "\r")
		sr = strings.Trim(sr, " ")
		expr := ""
		if sr == "S:" { // 代码区, 代码区仅提供行注释，注释可用# 分割
			for {
				i++
				sr = strings.Split(cmds[i], "#")[0]
				sr = strings.Trim(sr, "\r")
				sr = strings.Trim(sr, " ")
				if sr == ":E" {
					break
				}
				expr += sr + " " // 添加空格分割行
			}
		} else {
			expr = FindExpr(sr) // 任何非表达式都是注释
		}
		if expr != "" {
			c := Code{}
			// fmt.Println(expr)
			c.Init(expr)
			s := c.Read_Root()
			Eval(s, &EnvG)
		}
	}
}

func ExeIDLE() { // 解释执行
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("User>>")
		cmds, _, _ := reader.ReadLine()
		cmd := string(cmds)
		if cmd == "exit" {
			return
		} else {
			cmd = FindExpr(cmd)
			if cmd != "" {
				c := Code{}
				c.Init(cmd)
				s := c.Read_Root()
				fmt.Println(Eval(s, &EnvG))
			}
		}

	}
}

/*计算结束*/
func main() {
	args := os.Args
	if len(args) < 2 {
		ExeIDLE()
	} else {
		src, err := ioutil.ReadFile(args[1])
		if err == nil {
			ExeFile(string(src))
		} else {
			fmt.Println("打开文件出错！")
		}
	}
}
