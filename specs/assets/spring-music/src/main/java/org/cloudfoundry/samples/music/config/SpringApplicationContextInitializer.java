package org.cloudfoundry.samples.music.config;

import io.pivotal.cfenv.core.CfEnv;
import io.pivotal.cfenv.core.CfService;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.boot.autoconfigure.jdbc.DataSourceAutoConfiguration;
import org.springframework.context.ApplicationContextInitializer;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.core.env.ConfigurableEnvironment;
import org.springframework.core.env.MapPropertySource;
import org.springframework.core.env.Profiles;
import org.springframework.core.env.PropertySource;
import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.stream.Collectors;
import java.util.stream.Stream;

public class SpringApplicationContextInitializer implements ApplicationContextInitializer<ConfigurableApplicationContext> {

    private static final Log logger = LogFactory.getLog(SpringApplicationContextInitializer.class);

    private static final Map<String, List<String>> profileNameToServiceTags = new HashMap<>();
    static {
        profileNameToServiceTags.put("mysql", Collections.singletonList("mysql"));
    }

    @Override
    public void initialize(ConfigurableApplicationContext applicationContext) {
        ConfigurableEnvironment appEnvironment = applicationContext.getEnvironment();

        validateActiveProfiles(appEnvironment);

        addCloudProfile(appEnvironment);

        excludeAutoConfiguration(appEnvironment);
    }

    private void addCloudProfile(ConfigurableEnvironment appEnvironment) {
        CfEnv cfEnv = new CfEnv();

        List<String> profiles = new ArrayList<>();

        List<CfService> services = cfEnv.findAllServices();
        List<String> serviceNames = services.stream()
                .map(CfService::getName)
                .collect(Collectors.toList());

        logger.info("Found services " + StringUtils.collectionToCommaDelimitedString(serviceNames));

        for (CfService service : services) {
            for (String profileKey : profileNameToServiceTags.keySet()) {
                if (service.getTags().containsAll(profileNameToServiceTags.get(profileKey))) {
                    profiles.add(profileKey);
                }
            }
        }

        if (profiles.size() > 1) {
            throw new IllegalStateException(
                    "Only one service of the following types may be bound to this application: " +
                            profileNameToServiceTags.values().toString() + ". " +
                            "These services are bound to the application: [" +
                            StringUtils.collectionToCommaDelimitedString(profiles) + "]");
        }

        if (profiles.size() > 0) {
            logger.info("Setting service profile " + profiles.get(0));
            appEnvironment.addActiveProfile(profiles.get(0));
        }
    }

    private void validateActiveProfiles(ConfigurableEnvironment appEnvironment) {
        Set<String> validLocalProfiles = profileNameToServiceTags.keySet();

        List<String> serviceProfiles = Stream.of(appEnvironment.getActiveProfiles())
                .filter(validLocalProfiles::contains)
                .collect(Collectors.toList());

        if (serviceProfiles.size() > 1) {
            throw new IllegalStateException("Only one active Spring profile may be set among the following: " +
                    validLocalProfiles.toString() + ". " +
                    "These profiles are active: [" +
                    StringUtils.collectionToCommaDelimitedString(serviceProfiles) + "]");
        }
    }

    private void excludeAutoConfiguration(ConfigurableEnvironment environment) {
        List<String> exclude = new ArrayList<>();

        Map<String, Object> properties = Collections.singletonMap("spring.autoconfigure.exclude",
                StringUtils.collectionToCommaDelimitedString(exclude));

        PropertySource<?> propertySource = new MapPropertySource("springMusicAutoConfig", properties);

        environment.getPropertySources().addFirst(propertySource);
    }

    private void excludeDataSourceAutoConfiguration(List<String> exclude) {
        exclude.add(DataSourceAutoConfiguration.class.getName());
    }
}
