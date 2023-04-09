# PPL - Postgresql Procedural Language

PPL is a dialect of pl/pgsql with some edges taken off.

create table t (
   col1 integer,
   col2 text,
   col3 boolean
);

// set returning function
[volatile] function my_function2($arg integer): (a integer, b integer) {
}

// regular function
[volatile] function my_function($arg integer): integer {

    // results are available inside the { ... } as "col1", "col2":
    select * from t where col1  = $arg {
    }

    // values are selected into variables (maps to "select a,b,c strict into x,y,z")
    // x, y, z are automatically declared ? if we can.
    x, y, z := select col1, col2, col3 from t;

    // w is auto declared
    w := my_function(x, y, z);        // function calls don't need SELECT or PERFORM 
    v, w := my_function2(x, y, z);    // set-returning functions with Go-like syntax 
}