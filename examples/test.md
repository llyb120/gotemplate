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
select * from table1 {{ho}}
`122`
--# if !ho && test()
    --# use self: ho=ok, abc=ho_ho
      --# hook a
      hoho
      --# end
    --# end
--# else
    --# slot a
    hahah a
    --# end

    --# slot b
    hahah b
    --# end
--# end
```
