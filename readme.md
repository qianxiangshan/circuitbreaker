### 熔断器

1. 熔断器触发判定
2. 熔断器恢复判定
3. 熔断器报警.  



缺陷: c c从halfopen 状态到其他状态时,并不是唯一通过一个,而是一堆同时进行.能通过则过,不行就会进入close, too fast



https://yq.aliyun.com/articles/7443


Hystrix 
https://github.com/Netflix/Hystrix/wiki


record
total
success
fail
timeout
reject


open
half-open
close

alarm: close
rule: to open (timeout/all>30%||fail/all>50% )&&total>20
//next
matrix: costtime 50% 90% 




interface{
	run(context)error
	fallback(context)
}

