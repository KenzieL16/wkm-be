package repository

import (
	"database/sql"
	"fmt"
	"time"
	"wkm/entity"
	"wkm/request"
	"wkm/response"
	// "encoding/json"
)

type TicketSupportRepository interface {
	CreateTicketSupport(data request.TicketRequest, username string, tier uint32) (string, string, error)
	EditTicketSupport(noTicket string, data request.TicketRequest, username string) (string, error)
	ViewTicketSupport(noTicket string) (entity.TicketSupport, error)
	ListTicketUser(username string) ([]entity.TicketSupport, error)
	ListTicketIT(username string) ([]entity.TicketSupport, error)
	ListTicketQueue(month string, year string) ([]entity.TicketSupport, error)
}

type ticketSupportRepository struct {
	conn *sql.DB
}

func NewTicketSupportRepository(conn *sql.DB) TicketSupportRepository {
	return &ticketSupportRepository{
		conn: conn,
	}
}

func (ts *ticketSupportRepository) CreateTicketSupport(data request.TicketRequest, username string, tier uint32) (string, string, error) {
	// Generate nomor tiket
	noTicket, err := ts.GenerateNoTicket()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate no_ticket: %w", err)
	}

	// Dapatkan waktu Jakarta
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return "", "", fmt.Errorf("failed to load Jakarta timezone: %w", err)
	}
	created := time.Now().In(location)

	// Query untuk memasukkan data tiket
	query := `
		INSERT INTO ticket_support (
			no_ticket, kd_user, ` + "`case`" + `, status, 
			created, jenis_ticket, tier_ticket
		) VALUES (?, ?, ?, 2, ?, ?, ?)
	`
	_, err = ts.conn.Exec(
		query,
		noTicket,
		username, // Menggunakan username sebagai kd_user
		data.Case,
		created,
		data.JenisTicket,
		tier,
	)
	fmt.Println(data.Clients)
	if len(data.Clients) > 0 {
		for _, client := range data.Clients {
			_, err = ts.conn.Exec(`INSERT INTO ticket_client_ts (no_ticket, kd_user_client) VALUES (?,?)`, noTicket, client.Name)
			if err != nil {
				return "", "", fmt.Errorf("failed to insert ticket client: %w", err)
			}
		}
	}

	if err != nil {
		return "", "", fmt.Errorf("failed to insert ticket: %w", err)
	}

	result, err := ts.AssignTicket()
	if err != nil {
		return "", "", fmt.Errorf("failed to assign ticket: %w", err)
	}

	fmt.Printf("AssignTicket result: %v\n", result)

	// Jika IT support sedang penuh (result berisi "IT support sedang penuh"), tidak errorkan proses
	if result == "IT support sedang penuh" {
		// Log pesan, tapi tidak menghentikan pembuatan tiket
		fmt.Println("IT support sedang penuh, tiket tetap dibuat tanpa IT support.")
	}

	// Kembalikan nomor tiket yang baru saja dibuat
	return noTicket, result, nil
}

func (ts *ticketSupportRepository) AssignTicket() (string, error) {
	// Query untuk mencari IT support yang ON (status = 1)
	query := `
        SELECT kd_user
        FROM it_supports
        WHERE status = 0
        ORDER BY last_activity DESC
        LIMIT 1
    `
	var kdUser string

	err := ts.conn.QueryRow(query).Scan(&kdUser)

	if err != nil {
		if err == sql.ErrNoRows {
			// Jika tidak ada IT support yang tersedia (semua status = Off)
			fmt.Println("IT support sedang penuh (semua IT support OFF)")
			return "IT support sedang penuh", nil // Mengembalikan pesan bahwa tidak ada IT support yang tersedia
		}
		// Jika ada error lain saat query
		fmt.Printf("Error querying IT support: %v\n", err)
		return "", fmt.Errorf("failed to query IT support: %w", err)
	}

	fmt.Println("IT support found:", kdUser)

	// Lanjutkan dengan proses update tiket
	ticketQuery := `
        SELECT 
    t.no_ticket
    FROM 
    ticket_support t
    WHERE t.status = 2
        ORDER BY 
        t.tier_ticket, t.created
        LIMIT 1
    `
    var noTicket string
    row := ts.conn.QueryRow(ticketQuery)
    err = row.Scan(&noTicket)

    if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Tidak ada tiket")
			return "Tidak ada Tiket", nil
		}
		// Jika ada error lain saat query
		fmt.Printf("Error querying Ticket: %v\n", err)
		return "", fmt.Errorf("failed to query Ticket: %w", err)
	}


	now := time.Now()
	_, err = ts.conn.Exec(`
        UPDATE ticket_support
        SET kd_user_it = ?, assign_date = ?
        WHERE no_ticket = ?
    `, kdUser, now, noTicket)

	if err != nil {
		return "", fmt.Errorf("failed to update ticket_support: %w", err)
	}

	// Update it_supports
	_, err = ts.conn.Exec(`
        UPDATE it_supports
        SET last_activity = ?, status = 1
        WHERE kd_user = ?
    `, now, kdUser)

	if err != nil {
		return "", fmt.Errorf("failed to update it_supports: %w", err)
	}

	return "Tiket Sudah Masuk Antrian", nil
}

func (ts *ticketSupportRepository) GenerateNoTicket() (string, error) {
	// Dapatkan tahun dan bulan sekarang
	now := time.Now()
	year := now.Format("06")  // Format tahun dua digit (24 untuk 2024)
	month := now.Format("01") // Format bulan dua digit (12 untuk Desember)

	// Query untuk mendapatkan nomor urut terakhir berdasarkan tahun dan bulan
	query := `
		SELECT COALESCE(MAX(SUBSTRING(no_ticket, 7)), 0) + 1 AS next_number
		FROM ticket_support
		WHERE no_ticket LIKE ?
	`
	// Pola pencarian, contoh: "TK2412%"
	pattern := fmt.Sprintf("TK%s%s%%", year, month)

	var nextNumber int
	err := ts.conn.QueryRow(query, pattern).Scan(&nextNumber)
	if err != nil {
		return "", fmt.Errorf("failed to generate no_ticket: %w", err)
	}

	// Format nomor tiket, contoh: TK241200001
	noTicket := fmt.Sprintf("TK%s%s%05d", year, month, nextNumber)
	return noTicket, nil
}

func (ts *ticketSupportRepository) EditTicketSupport(noTicket string, data request.TicketRequest, username string) (string, error) {
	// Query untuk memperbarui data tiket
	query := `
        UPDATE ticket_support
    SET ` + "`case`" + ` = ?, jenis_ticket = ?, kd_user_it = ?, solution = ?, 
    assign_date = ?, finish_date = ?, 
    status = ?, modified = NOW(), modi_by = ?
    WHERE no_ticket = ?
    `
	var assignDate, finishDate *time.Time
	var status *int
	var kdUserIt *string
	var Solution *string

	if data.KdUserIt != "" {
		kdUserIt = &data.KdUserIt
		status = new(int)
		*status = 1
		finishDate = nil
		Solution = nil
	} else {
		kdUserIt = nil
		status = new(int)
		*status = 2
		finishDate = nil
		Solution = nil
	}
	if data.KdUserIt != "" {
		location, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			return "", fmt.Errorf("failed to load Jakarta timezone: %w", err)
		}
		now := time.Now().In(location)
		assignDate = &now
		finishDate = nil
		Solution = nil
		if data.Solution != "" {
			location, err := time.LoadLocation("Asia/Jakarta")
			if err != nil {
				return "", fmt.Errorf("failed to load Jakarta timezone: %w", err)
			}
			now := time.Now().In(location)
			finishDate = &now
			status = new(int)
			*status = 3
		}
	} else if data.Status == 0 || data.Status == 2 {
		assignDate = nil
		status = new(int)
		*status = 2
		finishDate = nil
	}

	if data.Status == 4 {
		status = &data.Status
	}

	_, err := ts.conn.Exec(
		query,
		data.Case,
		data.JenisTicket,
		kdUserIt,
		Solution,
		assignDate,
		finishDate,
		status,
		username,
		noTicket,
	)
	if err != nil {
		return "", fmt.Errorf("failed to update ticket: %w", err)
	}

	// Hapus data client yang lama
	_, err = ts.conn.Exec(`DELETE FROM ticket_client_ts WHERE no_ticket = ?`, noTicket)
	if err != nil {
		return "", fmt.Errorf("failed to delete old ticket client: %w", err)
	}

	// Tambahkan data client yang baru
	if len(data.Clients) > 0 {
		for _, client := range data.Clients {
			_, err = ts.conn.Exec(`INSERT INTO ticket_client_ts (no_ticket, kd_user_client) VALUES (?,?)`, noTicket, client.Name)
			if err != nil {
				return "", fmt.Errorf("failed to insert new ticket client: %w", err)
			}
		}
	}

	// Kembalikan nil jika sukses
	return noTicket, nil
}

func (ts *ticketSupportRepository) ViewTicketSupport(noTicket string) (entity.TicketSupport, error) {
	query := `
		SELECT 
    t.no_ticket, 
    u.name AS kd_user, 
    t.case, 
    t.status, 
    t.kd_user_it, 
    t.created, 
    t.modified, 
    t.modi_by, 
    t.assign_date, 
    t.finish_date, 
    t.jenis_ticket, 
    t.tier_ticket, 
    t.solution
    FROM 
        db_wkm.ticket_support t
    LEFT JOIN 
        users.mst_users u
    ON 
        t.kd_user = u.username 
    WHERE 
        t.no_ticket = ?
	`
	var ticket entity.TicketSupport
	err := ts.conn.QueryRow(query, noTicket).Scan(
		&ticket.NoTicket,
		&ticket.Kd_user,
		&ticket.Case,
		&ticket.Status,
		&ticket.KdUserIt,
		&ticket.Created,
		&ticket.Modified,
		&ticket.ModiBy,
		&ticket.AssignDate,
		&ticket.FinishDate,
		&ticket.JenisTicket,
		&ticket.TierTicket,
		&ticket.Solution,
	)
	if err != nil {
		return entity.TicketSupport{}, err
	}

	// Ambil data clients
	queryClients := `
		SELECT 
    		kd_user_client 
		FROM 
    		ticket_client_ts 
		WHERE 
    		no_ticket = ?
	`
	rows, err := ts.conn.Query(queryClients, noTicket)
	if err != nil {
		return entity.TicketSupport{}, err
	}
	defer rows.Close()

	var clients []response.TicketClient
	for rows.Next() {
		var kdUserClient *string
		err := rows.Scan(&kdUserClient)
		if err != nil {
			return entity.TicketSupport{}, err
		}

		// Menambahkan client ke dalam slice
		client := response.TicketClient{KdUserClient: kdUserClient}
		clients = append(clients, client)
	}
	ticket.Clients = clients
	return ticket, nil
}

func (ts *ticketSupportRepository) ListTicketUser(username string) ([]entity.TicketSupport, error) {
	// Query untuk mengambil tiket berdasarkan kd_user dan mengurutkannya
	query := `
		SELECT 
    t.no_ticket, 
    t.kd_user, 
    t.case, 
    t.status, 
    t.kd_user_it, 
    t.created, 
    t.modified, 
    t.modi_by, 
    t.assign_date, 
    t.finish_date, 
    t.jenis_ticket, 
    t.tier_ticket, 
    t.solution
FROM 
    ticket_support t
WHERE 
    t.kd_user = ?
ORDER BY 
    t.status, t.created;
	`

	// Menjalankan query
	rows, err := ts.conn.Query(query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Menyimpan hasil query ke dalam slice
	var tickets []entity.TicketSupport
	for rows.Next() {
		var ticket entity.TicketSupport
		err := rows.Scan(
			&ticket.NoTicket,
			&ticket.Kd_user,
			&ticket.Case,
			&ticket.Status,
			&ticket.KdUserIt,
			&ticket.Created,
			&ticket.Modified,
			&ticket.ModiBy,
			&ticket.AssignDate,
			&ticket.FinishDate,
			&ticket.JenisTicket,
			&ticket.TierTicket,
			&ticket.Solution,
		)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, ticket)
	}

	// Memeriksa apakah ada error saat membaca hasil
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tickets, nil
}

func (ts *ticketSupportRepository) ListTicketQueue(month string, year string) ([]entity.TicketSupport, error) {
    // Dasar query tanpa filter
    query := `
        SELECT 
            t.no_ticket, 
            t.kd_user, 
            t.case, 
            t.status, 
            t.kd_user_it, 
            t.created, 
            t.modified, 
            t.modi_by, 
            t.assign_date, 
            t.finish_date, 
            t.jenis_ticket, 
            t.tier_ticket, 
            t.solution
        FROM 
            ticket_support t
        WHERE 1=1
    `

    // Slice untuk menyimpan parameter query
    var args []interface{}

    // Tambahkan filter bulan jika tersedia
    if month != "" {
        query += " AND MONTH(t.created) = ?"
        args = append(args, month)
    }

    // Tambahkan filter tahun jika tersedia
    if year != "" {
        query += " AND YEAR(t.created) = ?"
        args = append(args, year)
    }

    // Tambahkan urutan
    query += `
        ORDER BY 
            t.status, t.tier_ticket, t.created
    `

    // Eksekusi query dengan parameter
    rows, err := ts.conn.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // Slice untuk menyimpan hasil query
    var tickets []entity.TicketSupport
    for rows.Next() {
        var ticket entity.TicketSupport
        err := rows.Scan(
            &ticket.NoTicket,
            &ticket.Kd_user,
            &ticket.Case,
            &ticket.Status,
            &ticket.KdUserIt,
            &ticket.Created,
            &ticket.Modified,
            &ticket.ModiBy,
            &ticket.AssignDate,
            &ticket.FinishDate,
            &ticket.JenisTicket,
            &ticket.TierTicket,
            &ticket.Solution,
        )
        if err != nil {
            return nil, err
        }
        tickets = append(tickets, ticket)
    }

    // Periksa error pada rows
    if err = rows.Err(); err != nil {
        return nil, err
    }

    return tickets, nil
}
	

func (ts *ticketSupportRepository) ListTicketIT(username string) ([]entity.TicketSupport, error) {
	// Query untuk mengambil tiket berdasarkan kd_user dan mengurutkannya
	query := `
		SELECT 
    t.no_ticket, 
    t.kd_user, 
    t.case, 
    t.status, 
    t.kd_user_it, 
    t.created, 
    t.modified, 
    t.modi_by, 
    t.assign_date, 
    t.finish_date, 
    t.jenis_ticket, 
    t.tier_ticket, 
    t.solution
FROM 
    ticket_support t
WHERE 
    t.kd_user_it = ?
ORDER BY 
    t.status, t.tier_ticket, t.created;
	`

	// Menjalankan query
	rows, err := ts.conn.Query(query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Menyimpan hasil query ke dalam slice
	var tickets []entity.TicketSupport
	for rows.Next() {
		var ticket entity.TicketSupport
		err := rows.Scan(
			&ticket.NoTicket,
			&ticket.Kd_user,
			&ticket.Case,
			&ticket.Status,
			&ticket.KdUserIt,
			&ticket.Created,
			&ticket.Modified,
			&ticket.ModiBy,
			&ticket.AssignDate,
			&ticket.FinishDate,
			&ticket.JenisTicket,
			&ticket.TierTicket,
			&ticket.Solution,
		)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, ticket)
	}

	// Memeriksa apakah ada error saat membaca hasil
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tickets, nil
}