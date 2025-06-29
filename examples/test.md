# test

## sql1
```sql
select * from table1
where 
1 = 1
AND id = 1,   --# val x if true
AND id = 2,   --# val x? 
and b in ('') --# each arr by ''
and c = 3     --# abc?
and d = 4     --# a by 4

--# use test2
--# end

--# use foo.sql3
--# end
我是分割线

--# use sql2: a=3,b=4
  --# hook a
  纳尼 --# if b
  --# if 1 > 0
  我是替换的slot
  --# else 
  4343
  --# end
  --# end
--# end

before hook


    
```

## sql2
```sql
haha

--# slot a
123
--# end
```

## test2
```sql
select * from table1
where 
--# trim and safe 1 < 0
--# for key, value := range mp
  and {{key}} = {{value}}
--# end
--# end
```


## test3
测试自引用
```sql
--# Test()

select * from table1 {{ho}}
`122` --# testfn()? by 1
and id in ('1', "2", 123) and id not in (1,2,3) --# each Items $$ each Items2
1, --# val a
and c = 1 and d = 234 --# val a $$ val b
and c = '1' and d = '234' --# val a $$ val b
and c = 1 and d = '234' --# val a $$ val b
--# if !ho 
    --# use self: ho=ok, abc=ho_ho 
      --# hook a
      hoho redo! --# if dore
      hoho --# if !dore
      --# end
    --# end
--# else
    --# slot a
    hahah a redo! --# if dore
    hahah a --# if !dore
    --# end

    --# use test_use_current_context: context=current2
    --# end

    --# slot b
    hahah b
    --# end
--# end

--# redo a: dore=true
```

## test_use_current_context
```sql
  it will be shown
  --# redo a
     
    --# slot ccc
     i am ccc
    --# end
     
    --# redo ccc if true
    --# redo ccc if false
```


## test_run_code
```go prev
for i, v := range _params {
    if v == Date {
        _params[i] = PrevDate
    }
}
echo(_sql, _params)
```
```sql
select
--# slot a
case when date > '2025-01-01' then '1' else '2' end as a --# val Date by '2025-01-01'
--# end
--# prev()
from ok
```