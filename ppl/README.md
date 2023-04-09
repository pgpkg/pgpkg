# PPL - Postgresql Procedural Language

PPL - proper name TBA - is a dialect of pl/pgsql with some edges taken off.
The idea is to create a transpiler to plpgsql. These are just some scratchings
of an idea at the moment.


create table t (
   col1 integer,
   col2 text,
   col3 boolean
);

// record- returning function
function my_function2($arg integer): (a integer, b integer) {
}

// setof-record- returning function
function my_function2($arg integer): (a integer, b integer)[] {
}

// regular function. args are always prefixed with $ to avoid name clashes
// with database objects. note: quoted arg names work in PG functions, which
// would let us transpile them.
function my_function($arg integer): integer {

    // results are available inside the { ... } as "col1", "col2"
    // (it's unclear how to transpile this...)
    select * from t where col1 = $arg {
    }

    // values are selected into variables (maps to "select a,b,c strict into x,y,z")
    // x, y, z are automatically declared ? if we can, like in Go.
    $x, $y, $z := select col1, col2, col3 from t;

   // we might need to include data types in declarations, so instead:
   $x: integer, $y: text, $z: boolean := select col1, col2, col3 from t;

    // w is auto declared
    $w := my_function(x, y, z);        // function calls don't need SELECT or PERFORM 
    $v, $w := my_function2(x, y, z);    // set-returning functions with Go-like syntax 
}

// Note that we could store a line number map in the comments for the sql function, which would
// let us print stack traces in ppl.

// function modifiers at front:

immutable function f(): integer { return 1; } 