package projectionfreshness

import "time"

func buildFreshnessComponents(
	metrics freshnessMetrics,
	config Config,
) []Component {
	recentSupportScore := clampUnit(
		float64(metrics.recentCount) / float64(config.TargetRecentNeighborCount),
	)
	return []Component{
		{Name: ComponentNewestAge, Score: ageScore(metrics.newestAge, config.MaximumNewestNeighborAge), Weight: config.NewestAgeWeight},
		{Name: ComponentMeanAge, Score: ageScore(metrics.meanAge, config.MaximumMeanNeighborAge), Weight: config.MeanAgeWeight},
		{Name: ComponentOldestAge, Score: ageScore(metrics.oldestAge, config.MaximumOldestNeighborAge), Weight: config.OldestAgeWeight},
		{Name: ComponentRecentSupport, Score: recentSupportScore, Weight: config.RecentSupportWeight},
	}
}

func weightedComponentScore(components []Component) float64 {
	score := 0.0
	for _, component := range components {
		score += component.Score * component.Weight
	}
	return clampUnit(score)
}

func ageScore(age time.Duration, maximum time.Duration) float64 {
	if maximum <= 0 || age < 0 {
		return 0
	}
	return clampUnit(1 - float64(age)/float64(maximum))
}
