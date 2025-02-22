func BenchmarkWordOperations(b *testing.B) {
    db := setupTestDB()
    defer db.Close()

    // Single word creation benchmark
    b.Run("create_single", func(b *testing.B) {
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            word := models.Word{
                Japanese: "猫",
                Romaji:   "neko",
                English:  fmt.Sprintf("cat_%d", i),
            }
            createWord(db, word)
        }
    })

    // Bulk word creation benchmark
    b.Run("create_bulk", func(b *testing.B) {
        words := make([]models.Word, 100)
        for i := range words {
            words[i] = models.Word{
                Japanese: fmt.Sprintf("猫%d", i),
                Romaji:   fmt.Sprintf("neko%d", i),
                English:  fmt.Sprintf("cat_%d", i),
            }
        }
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            createBulkWords(db, words)
        }
    })

    // Read operation benchmark
    b.Run("read", func(b *testing.B) {
        word := models.Word{
            Japanese: "犬",
            Romaji:   "inu",
            English:  "dog",
        }
        id := createWord(db, word)
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            getWord(db, id)
        }
    })

    // Search operation benchmark
    b.Run("search", func(b *testing.B) {
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            searchWords(db, "cat", 10)
        }
    })
}