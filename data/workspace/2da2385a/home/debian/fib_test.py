#!/usr/bin/env python3
"""
斐波那契数列计算脚本
计算并打印斐波那契数列的前20项
定义：F(1)=1, F(2)=1, F(n)=F(n-1)+F(n-2) (n>=3)
"""


def fibonacci(n):
    """计算斐波那契数列前n项，返回列表"""
    if n <= 0:
        return []
    if n == 1:
        return [1]

    # 初始化前两项
    fib = [1, 1]

    # 从第3项开始迭代计算
    for i in range(2, n):
        fib.append(fib[i - 1] + fib[i - 2])

    return fib


def main():
    n = 20
    result = fibonacci(n)

    print(f"斐波那契数列前{n}项：")
    for i, value in enumerate(result, start=1):
        print(f"F({i}) = {value}")


if __name__ == "__main__":
    main()
