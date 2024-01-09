package org.cloudfoundry.samples.music;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.actuate.health.Health;
import org.springframework.boot.actuate.health.HealthAggregator;
import org.springframework.boot.actuate.health.OrderedHealthAggregator;
import org.springframework.core.env.Environment;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

import static java.util.stream.Collectors.toMap;


/**
 * Health Aggregator class that only considers the persistence provider that is used.
 * For example if H2 is used then it will not aggregate the health of mysql
 */
@Component
public class SpringMusicHealthAggregator implements HealthAggregator   {

    private final List<String> excluded = new ArrayList<>();
    private final OrderedHealthAggregator orderedHealthAggregator = new OrderedHealthAggregator();

    @Autowired
    public SpringMusicHealthAggregator(Environment environment) { }

    @Override
    public Health aggregate(Map<String, Health> healths){

       Map<String,Health>  filtered = healths.entrySet().stream().filter( entry -> !excluded.contains(entry.getKey())).
                collect(toMap(entry -> entry.getKey(), entry -> entry.getValue()));

       Health filteredHealth = orderedHealthAggregator.aggregate(filtered);
       return filteredHealth;
    }

}
