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

## sql2
```sql
select * from table1
where 

--# use foo.sql3

    
```

## testSql
```sql
select * from abc where
    1 = 1
--# trim and     safe 123
and a = 1    --# when 1 < 0
and b = 1   --# when 1 < 0
and c = 1    --# when 1 < 0
and d = 1     --# when 1 < 0
--#end

```

## sql4
```sql
update table1 set
--# for key, value := range mp  
    , 1231
--# end 
```


## test2
```sql
select * from table1
where 
1 = 1
--# for key, value := range mp
  and {{key}} = {{value}}
--# end
```
