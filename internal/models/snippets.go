package models

import (
	"database/sql"
	"errors"
	"time"
)

// Define a Snippet type to hold the data for an individual snippet. Notice how
// the fields of the struct correspond to the fields in our MySQL snippets
// table?

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// Define a SnippetModel type which wraps a sql.DB connection pool.

type SnippetModel struct {
	DB *sql.DB
}

// This will insert a new snippet into the database.

func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	//write the sql statement we want to execute.I've split it over two lines
	// for readability (w (which is why its surrounded with backquotes instead
	//of normal double quotes)
	stmt := `INSERT INTO snippets (title, content, created, expires)
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`
	//Use the Exec() method on the embedded connection pool to execute the
	// statement. The first parameter is the SQL statement, followed by the
	// title, content and expiry values for the placeholder parameters. This
	// method returns a sql.Result type, which contains some basic
	// information about what happened when the statement was executed.
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}
	// Use the LastInsertId() method on the result to get the ID of our
	// newly inserted record in the snippets table.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	// The ID returned has the type int64, so we convert it to an int type
	// before returning.
	return int(id), nil
}

// This will return a specific snippet based on its id.

func (m *SnippetModel) Get(id int) (*Snippet, error) {
	//	write the sql statement we want to excute. Again, I've split it over two
	//	lines for readability
	stmt := `SELECT id, title, content, created, expires FROM snippets WHERE expires > UTC_TIMESTAMP() and id = ?`
	// use the QueryRow() method on the connection pool to execute
	// Our SQL statement, passing in the untrusted id variable as the value for the
	// placeholder parameter. this return a pointer to a sql.Row object which
	// holds the result from database.
	//row := m.DB.QueryRow(stmt, id)
	// Initialize a pointer to a new zeroed Snippet structure
	snippet := &Snippet{}
	// use row.Scan() to copy the values from each field in sql.Row to the corresponding
	// field in the Snippet struct. Notice that arguments, to row.Scan are *pointers* to place you want
	// to copy the data into, and the number of arguments be exactly the sames as the number of
	// columns returned by your statement.
	//err := row.Scan(&snippet.ID, &snippet.Title, &snippet.Content, &snippet.Created, &snippet.Expires)
	// new code:
	err := m.DB.QueryRow(stmt, id).Scan(&snippet.ID, &snippet.Title, &snippet.Content, &snippet.Created, &snippet.Expires)
	if err != nil {
		// if the query return no rows, then row.Scan() will return a
		// sql.ErrorNoRows error. we use the errors.Is() function check for that
		// error specifically, and return our own ErrorNoRecord error
		// instead ( we'll create this a moments)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	// if everything went OK  then return the snippet object
	return snippet, nil
}

// This will return the 10 most recently created snippets.

func (m *SnippetModel) Latest() ([]*Snippet, error) {
	// write sql statement we want to execute
	stmt := `SELECT id, title, content, created, expires FROM snippets 
	WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`
	// Use the Query() method on the connection pool to execute our
	// SQL statement. This return a sql.Rows result set containing the result
	// of our query
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	// We defer rows.Close() to ensure the sql.Rows result set is
	// always property closed before the Latest() method return.
	// this defer statement should come *after* you check for an error from the Query()
	// method. Otherwise, if Query() return an error, you'll get a panic
	// trying to close a nil result set
	defer rows.Close()
	// Initialize an empty slice to hold the snippet struct
	snippets := []*Snippet{}
	// Use rows.Next to iterate through the rows in result set.
	// this prepares the first (and then each subsequent) row to be acted on by the
	// rows.Scan() method. If iteration over all the rows completes then the
	// result set automatically closes itself and frees-up the underlying database connection.
	for rows.Next() {
		snippet := &Snippet{}
		// Use row.Scan() to copy the values from each field in the row to the
		// new Snippet object that we created. Again, the arguments to row.Scan()
		// must be pointers to the place you want to copy the data into, and the
		// number of arguments must be exactly the same as the number of
		// columns returned by your statement.
		err = rows.Scan(&snippet.ID, &snippet.Title, &snippet.Content, &snippet.Created, &snippet.Expires)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, snippet)

	}
	// when the rows.Next() loop has the finished we call a rows.Err() to retrieve any
	// error that was encountered during the iteration. It's important to call
	// this - don't assume that a successful iteration was completed
	// over the whole result set.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// if everything went ok then return the snippets slice
	return snippets, nil
}
