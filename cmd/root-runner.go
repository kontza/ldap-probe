package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-ldap/ldap/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func rootRunner(cmd *cobra.Command, args []string) {
	showFull := viper.GetBool("full")
	verbose := viper.GetBool("verbose")
	if verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Figure out the password
	adPassStr := viper.GetString("ad-password")
	if strings.TrimSpace(adPassStr) == "" {
		log.Info().Msg("No password found in configuration, trying out $HOME/.adpass")
		adPass, err := os.ReadFile(os.ExpandEnv("$HOME/.adpass"))
		if err != nil {
			log.Fatal().Err(err).Msg("ReadFile failed due to")
		}
		adPassStr = string(adPass)
	}

	dialUrl := viper.GetString("dial-url")
	log.Info().Str("dial URL", dialUrl).Msg("Configured")
	l, err := ldap.DialURL(dialUrl)
	if err != nil {
		log.Fatal().Err(err).Msg("DialURL failed due to")
	}
	defer l.Close()

	bindDn := viper.GetString("bind-dn")
	log.Info().Str("bind DN", bindDn).Msg("Configured")
	err = l.Bind(bindDn, adPassStr)
	if err != nil {
		log.Fatal().Err(err).Msg("Bind failed due to")
	}

	baseDn := viper.GetString("base-dn")
	log.Info().Str("base DN", baseDn).Msg("Configured")
	for _, searchTerm := range args {
		subQuery := fmt.Sprintf("(&(objectclass=user)(|(sAMAccountName=%s)(userPrincipalName=%s)))", searchTerm, searchTerm)
		log.Info().Str("sub query", subQuery).Msg("Using")
		searchRequest := ldap.NewSearchRequest(
			baseDn,
			ldap.ScopeWholeSubtree,
			ldap.NeverDerefAliases,
			0,
			0,
			false,
			subQuery,
			[]string{"dn", "cn", "phone"},
			nil,
		)

		if showFull {
			searchRequest.Attributes = []string{"*"}
		}

		if verbose {
			log.Debug().Strs("Attributes", searchRequest.Attributes).Msg("Search")
			l.Debug.Enable(true)
		}

		sr, err := l.Search(searchRequest)
		if err != nil {
			log.Fatal().Err(err).Msg("Search failed due to")
		}

		for _, entry := range sr.Entries {
			log.Info().Msg("Result:")
			log.Info().Str("  DN", entry.DN).Send()
			for _, attr := range entry.Attributes {
				prefix := log.Info().Interface("  Name", attr.Name)
				if len(attr.Values) > 1 {
					prefix.Str("Values", "...").Send()
					for _, value := range attr.Values {
						log.Info().Interface("    ", value).Send()
					}
				} else {
					prefix.Interface("Values", attr.Values).Send()
				}
			}
		}
	}
}
