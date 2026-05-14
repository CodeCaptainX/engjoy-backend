package utils

// import (
// 	"crypto/rand"
// 	"database/sql/driver"
// 	"fmt"
// 	"os"
// 	"reflect"

// 	"scratch_card_admin/pkg/constant"
// 	pkg_model "scratch_card_admin/pkg/models"
// 	"strconv"
// 	"strings"

// 	go_json "github.com/goccy/go-json"
// 	"github.com/google/uuid"
// 	"github.com/jmoiron/sqlx"
// 	"github.com/oklog/ulid"

// 	"github.com/gofiber/fiber/v2"
// )

// func JSONToMap(jsonStr string) (map[string]interface{}, error) {
// 	var dataMap map[string]interface{}
// 	err := go_json.Unmarshal([]byte(jsonStr), &dataMap)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return dataMap, nil
// }

// // JSONToStruct is a generic function that converts a JSON string to a Go struct.
// // It accepts any type T as a generic argument and returns an instance of T or an error if unmarshalling fails.
// func JSONToStruct[T any](jsonString string) (T, error) {
// 	var result T
// 	err := go_json.Unmarshal([]byte(jsonString), &result)
// 	if err != nil {
// 		return result, err
// 	}
// 	return result, nil
// }

// func ConvertStringToNumber(s string) int {
// 	if n, err := strconv.Atoi(s); err == nil {
// 		return n
// 	}
// 	return 0
// }

// func TrimSpaces(values interface{}) interface{} {
// 	// Check if values is a slice or array
// 	val := reflect.ValueOf(values)
// 	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
// 		// Return the input value itself if it's not a slice or array
// 		return values
// 	}

// 	// Initialize a new slice to store trimmed values
// 	trimmedValues := make([]interface{}, val.Len())

// 	// Iterate through each element of the slice or array
// 	for i := 0; i < val.Len(); i++ {
// 		elem := val.Index(i)

// 		// Check if the element is a string
// 		if elem.Kind() == reflect.String {
// 			// Trim spaces from the string element
// 			trimmedValues[i] = strings.TrimSpace(elem.String())
// 		} else {
// 			// For non-string elements, convert them to interface{}
// 			trimmedValues[i] = elem.Interface()
// 		}
// 	}

// 	return trimmedValues
// }

// // ScanJSON scans a JSON-encoded value into the provided destination struct.
// func ScanJSON(value interface{}, dest interface{}) error {
// 	bytes, ok := value.([]byte)
// 	if !ok {
// 		return fmt.Errorf("failed to scan: expected []byte, got %T", value)
// 	}

// 	// Unmarshal the JSON data into the provided destination struct
// 	if err := go_json.Unmarshal(bytes, dest); err != nil {
// 		return fmt.Errorf("failed to unmarshal JSON: %w", err)
// 	}
// 	return nil
// }

// // ValueJSON serializes the given source interface into a JSON byte slice,
// // and returns it as a driver.Value for database storage or transmission.
// func ValueJSON(src interface{}) (driver.Value, error) {
// 	return go_json.Marshal(src)
// }

// func GetClientInfo(c *fiber.Ctx) *pkg_model.ClientInfoResponse {
// 	userAgent := c.Get("User-Agent")

// 	// Get IP safely
// 	clientIP := c.IP()
// 	if clientIP == "" {
// 		clientIP = "0.0.0.0" // default IP to prevent inet errors
// 	}

// 	// Determine OS directly
// 	operatingSystem := "unknown"
// 	if userAgent != "" {
// 		switch {
// 		case strings.Contains(userAgent, "Windows"):
// 			operatingSystem = "Windows"
// 		case strings.Contains(userAgent, "Macintosh"):
// 			operatingSystem = "macOS"
// 		case strings.Contains(userAgent, "Linux"):
// 			operatingSystem = "Linux"
// 		case strings.Contains(userAgent, "Android"):
// 			operatingSystem = "Android"
// 		case strings.Contains(userAgent, "iPhone"), strings.Contains(userAgent, "iPad"):
// 			operatingSystem = "iOS"
// 		}
// 	}

// 	return &pkg_model.ClientInfoResponse{
// 		UserAgent:       userAgent,
// 		Ip:              clientIP,
// 		OperatingSystem: operatingSystem,
// 	}
// }

// // GetEnv retrieves the value of the environment variable named by the key.
// // If the variable is not set, it returns the provided default value.
// func GetEnv(key string) string {
// 	if value, exists := os.LookupEnv(key); exists {
// 		return value
// 	}
// 	return ""
// }

// func GetenvInt(key string, defaultValue int) int {
// 	valueStr := os.Getenv(key)
// 	value, err := strconv.Atoi(valueStr)
// 	if err != nil {
// 		return defaultValue
// 	}
// 	return value
// }

// func GenerateULID() ulid.ULID {
// 	t := constant.CurrentTime().UTC() // Get the current UTC time
// 	entropy := rand.Reader            // Use crypto/rand for random entropy
// 	// Generate the ULID using the timestamp and entropy
// 	return ulid.MustNew(ulid.Timestamp(t), entropy)
// }

// func Contains[T comparable](arr []T, value T) bool {
// 	for _, item := range arr {
// 		// Convert both item and value to string for case-insensitive comparison

// 		if strings.EqualFold(fmt.Sprintf("%v", item), fmt.Sprintf("%v", value)) {
// 			return true
// 		}
// 	}
// 	return false
// }

// // IsValidUUID validates whether a string is a valid UUID
// func IsValidUUID(u string) bool {
// 	_, err := uuid.Parse(u)
// 	return err == nil
// }

// func IsValidGameType(gameType string) bool {
// 	validTypes := map[string]bool{
// 		constant.CARD_LINE_2S1W:                        true,
// 		constant.CARD_ZHANG_FEI_1S1W:                   true,
// 		constant.CARD_FARMERS_LOOKING_FOR_ANIMALS_1S1W: true,
// 		constant.HUNTING_SHOOTING_ANIMALS_1S1W:         true,
// 		constant.CARD_DREAM_1S1W:                       true,
// 		constant.CARD_CRAP_1S1W:                        true,
// 		constant.CARD_LINE_JACKPOT_777_1S1W:            true,
// 	}
// 	return validTypes[gameType]
// }

// func GetNextSequenceID(q sqlx.Ext, sequenceName string) (int64, error) {
// 	var nextID int64

// 	query := fmt.Sprintf(`SELECT nextval('%s')`, sequenceName)

// 	err := sqlx.Get(q, &nextID, query)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to get nextval from sequence %s: %w", sequenceName, err)
// 	}

// 	return nextID, nil
// }

// func GetNextSeqId(seqName string, db interface{}) (int64, error) {

// 	var nextValue int64
// 	query := fmt.Sprintf("SELECT nextval('%s')", seqName)

// 	switch conn := db.(type) {
// 	case *sqlx.DB:
// 		if err := conn.QueryRow(query).Scan(&nextValue); err != nil {
// 			return 0, fmt.Errorf("failed to get next sequence value for %s: %w", seqName, err)
// 		}
// 	case *sqlx.Tx:
// 		if err := conn.QueryRow(query).Scan(&nextValue); err != nil {
// 			return 0, fmt.Errorf("failed to get next sequence value for %s: %w", seqName, err)
// 		}
// 	default:
// 		return 0, fmt.Errorf("unsupported db type %T", db)
// 	}

// 	// fmt.Println("🚀 ~ file: db.go ~ line 29 ~ funcGetNextSeqId ~ nextValue : ", nextValue)
// 	return nextValue, nil
// }
