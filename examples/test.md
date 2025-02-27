# test

## sql1
```sql
select * from table1
where 
1 = 1
AND id = 1,   --# val x when true
AND id = 2,   --# val x? 
and b in ('') --# each arr by ''
and c = 3     --# abc?
and d = 4     --# a by 4

--# use test2

--# use foo.sql3

我是分割线

--# use sql2 as u2

before hook

--# hook u2.a
纳尼 --# when b
--# if 1 > 0
我是替换的slot
--# end
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
